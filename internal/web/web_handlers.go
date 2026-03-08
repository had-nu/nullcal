package web

import (
	"bufio"
	"crypto/sha1" //nolint:gosec // required by RFC 6455
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/had-nu/nullcal/internal/store"
)

// client represents a single browser connection.
// Messages are queued on the send channel and written by writePump.
type client struct {
	hub  *Hub
	conn io.ReadWriteCloser
	send chan []byte
}

// inboundMsg is the envelope the browser sends for any action.
type inboundMsg struct {
	Type   string           `json:"type"`   // "create" | "update" | "delete" | "setStatus"
	Task   *taskDTO         `json:"task"`   // for create / update
	ID     string           `json:"id"`     // for delete / setStatus
	Status store.TaskStatus `json:"status"` // for setStatus
}

// taskDTO is the shape the browser uses to describe a task on create/update.
type taskDTO struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	ProjectTag  string  `json:"project_tag"`
	DueAt       *string `json:"due_at"` // "YYYY-MM-DD" or null
}

// handleWS upgrades the connection to WebSocket without external dependencies,
// implementing the handshake per RFC 6455 §4.2.2.
func (h *Hub) handleWS(w http.ResponseWriter, r *http.Request) {
	if !isWebSocketUpgrade(r) {
		http.Error(w, "expected WebSocket upgrade", http.StatusBadRequest)
		return
	}

	key := r.Header.Get("Sec-Websocket-Key")
	if key == "" {
		http.Error(w, "missing Sec-WebSocket-Key", http.StatusBadRequest)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "hijacking not supported", http.StatusInternalServerError)
		return
	}

	conn, buf, err := hj.Hijack()
	if err != nil {
		slog.Error("hijack failed", "err", err)
		return
	}

	accept := wsAcceptKey(key)
	resp := "HTTP/1.1 101 Switching Protocols\r\n" +
		"Upgrade: websocket\r\n" +
		"Connection: Upgrade\r\n" +
		"Sec-WebSocket-Accept: " + accept + "\r\n\r\n"

	if _, err := fmt.Fprint(conn, resp); err != nil {
		_ = conn.Close()
		slog.Error("writing handshake", "err", err)
		return
	}

	c := &client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 32),
	}
	h.register(c)

	// Send initial state immediately after handshake.
	if err := h.broadcast(); err != nil {
		slog.Error("initial broadcast failed", "err", err)
	}

	go c.writePump()
	c.readPump(buf)
}

// isWebSocketUpgrade checks the essential upgrade headers.
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket") &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// wsAcceptKey computes the Sec-WebSocket-Accept value per RFC 6455 §4.2.2.
func wsAcceptKey(clientKey string) string {
	const magic = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New() //nolint:gosec // SHA-1 is mandated by the WebSocket RFC, not for security
	_, _ = h.Write([]byte(clientKey + magic))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// readPump reads frames from the browser, decodes action messages,
// executes mutations against the store, and triggers a broadcast.
func (c *client) readPump(buf *bufio.ReadWriter) {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()

	for {
		_ = c.conn.(interface{ SetDeadline(time.Time) error }).SetDeadline( //nolint:errcheck
			time.Now().Add(readDeadline),
		)

		payload, err := wsReadFrame(buf.Reader)
		if err != nil {
			// Normal close or network error — no log noise.
			return
		}

		if len(payload) > maxMessageLen {
			slog.Warn("message too large, dropping", "len", len(payload))
			continue
		}

		var msg inboundMsg
		if err := json.Unmarshal(payload, &msg); err != nil {
			slog.Warn("invalid JSON from client", "err", err)
			continue
		}

		if err := c.hub.handleAction(msg); err != nil {
			slog.Error("handling action", "type", msg.Type, "err", err)
			continue
		}

		if err := c.hub.broadcast(); err != nil {
			slog.Error("broadcast after action", "err", err)
		}
	}
}

// writePump drains the send channel and writes frames to the browser.
func (c *client) writePump() {
	for msg := range c.send {
		if err := wsWriteFrame(c.conn, msg); err != nil {
			return
		}
	}
}

// handleAction dispatches a client message to the appropriate store method.
func (h *Hub) handleAction(msg inboundMsg) error {
	switch msg.Type {
	case "create":
		if msg.Task == nil {
			return fmt.Errorf("create: missing task payload")
		}
		t, err := dtoToTask(msg.Task)
		if err != nil {
			return fmt.Errorf("create: %w", err)
		}
		return h.store.CreateTask(&t)

	case "update":
		if msg.Task == nil {
			return fmt.Errorf("update: missing task payload")
		}
		t, err := dtoToTask(msg.Task)
		if err != nil {
			return fmt.Errorf("update: %w", err)
		}
		return h.store.UpdateTask(&t)

	case "delete":
		if msg.ID == "" {
			return fmt.Errorf("delete: missing id")
		}
		return h.store.DeleteTask(msg.ID)

	case "setStatus":
		if msg.ID == "" {
			return fmt.Errorf("setStatus: missing id")
		}
		return h.store.SetTaskStatus(msg.ID, msg.Status)

	default:
		return fmt.Errorf("unknown message type: %q", msg.Type)
	}
}

// dtoToTask converts a taskDTO from the browser into a store.Task.
func dtoToTask(d *taskDTO) (store.Task, error) {
	t := store.Task{
		ID:          d.ID,
		Title:       strings.TrimSpace(d.Title),
		Description: strings.TrimSpace(d.Description),
		ProjectTag:  strings.TrimSpace(d.ProjectTag),
		Status:      store.TaskStatusBacklog,
	}
	if t.Title == "" {
		return store.Task{}, fmt.Errorf("title is required")
	}
	if d.DueAt != nil && *d.DueAt != "" {
		due, err := time.Parse("2006-01-02", *d.DueAt)
		if err != nil {
			return store.Task{}, fmt.Errorf("invalid due_at format (expected YYYY-MM-DD): %w", err)
		}
		t.DueAt = &due
	}
	return t, nil
}

// handleIndex serves the embedded HTML dashboard.
func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(indexHTML))
}

// handleFavicon serves the embedded SVG application icon.
func handleFavicon(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(iconSVG))
}

// ── Minimal RFC 6455 frame codec ─────────────────────────────────────────────
// We only handle text frames (opcode 0x1) and connection-close (0x8).
// The browser always masks client→server frames; we never mask server→client.

// wsReadFrame reads one WebSocket frame from r and returns the unmasked payload.
func wsReadFrame(r *bufio.Reader) ([]byte, error) {
	// Byte 0: FIN + opcode
	b0, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	opcode := b0 & 0x0F
	if opcode == 0x8 { // Close frame
		return nil, io.EOF
	}

	// Byte 1: MASK bit + payload length
	b1, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	masked := (b1 & 0x80) != 0
	length := int(b1 & 0x7F)

	switch length {
	case 126:
		var ext [2]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return nil, err
		}
		length = int(ext[0])<<8 | int(ext[1])
	case 127:
		var ext [8]byte
		if _, err := io.ReadFull(r, ext[:]); err != nil {
			return nil, err
		}
		length = int(ext[4])<<24 | int(ext[5])<<16 | int(ext[6])<<8 | int(ext[7])
	}

	var maskKey [4]byte
	if masked {
		if _, err := io.ReadFull(r, maskKey[:]); err != nil {
			return nil, err
		}
	}

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	if masked {
		for i := range payload {
			payload[i] ^= maskKey[i%4]
		}
	}

	return payload, nil
}

// wsWriteFrame writes a single unmasked text frame to w.
func wsWriteFrame(w io.Writer, payload []byte) error {
	length := len(payload)
	var header []byte

	switch {
	case length <= 125:
		header = []byte{0x81, byte(length)} //nolint:gosec
	case length <= 65535:
		header = []byte{0x81, 126, byte(length >> 8), byte(length)} //nolint:gosec
	default:
		header = []byte{ //nolint:gosec
			0x81, 127,
			0, 0, 0, 0,
			byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(length), //nolint:gosec
		}
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
