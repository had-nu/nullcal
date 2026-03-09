// Package web provides a browser-based UI for nullcal over WebSocket.
//
// The server maintains a Hub that tracks connected clients. Any mutation
// (create, update, delete, set status) updates SQLite and broadcasts the
// complete current state to all clients — the browser is always a pure
// reflection of the database, never the source of truth.
package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
	"github.com/had-nu/nullcal/internal/sync/gcal"
)

const (
	writeTimeout  = 10 * time.Second
	readDeadline  = 60 * time.Second
	maxMessageLen = 4096
)

// Hub tracks all active WebSocket connections and serialises broadcasts.
// All exported methods are safe for concurrent use.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}

	store  *store.Store
	config *config.Config
	gcal   *gcal.Adapter // nil when GCal sync is not enabled
}

// NewHub creates a Hub ready to accept connections.
func NewHub(s *store.Store, cfg *config.Config) *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
		store:   s,
		config:  cfg,
	}
}

// SetGCal attaches an authenticated GCal adapter for bidirectional sync.
func (h *Hub) SetGCal(a *gcal.Adapter) { h.gcal = a }

// register adds a client to the hub.
func (h *Hub) register(c *client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// unregister removes a client and closes its send channel.
func (h *Hub) unregister(c *client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
}

// broadcast serialises the current DB state and sends it to every client.
// Clients that are too slow (send channel full) are dropped.
func (h *Hub) broadcast() error {
	msg, err := h.buildStateMessage()
	if err != nil {
		return fmt.Errorf("building state message: %w", err)
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			// Client is not keeping up — drop it.
			go h.unregister(c)
		}
	}
	return nil
}

// buildStateMessage assembles the full current state as a JSON payload.
func (h *Hub) buildStateMessage() ([]byte, error) {
	tasks, err := h.store.ListTasks()
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}

	calEvents, err := h.store.ListCalendarEvents()
	if err != nil {
		return nil, fmt.Errorf("listing calendar events: %w", err)
	}

	type routineBlockDTO struct {
		Weekday   int    `json:"weekday"` // 0=Sun … 6=Sat
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Label     string `json:"label"`
		Tag       string `json:"project_tag"`
	}

	type calEventDTO struct {
		ExternalID  string `json:"external_id"`
		Source      string `json:"source"`
		Title       string `json:"title"`
		StartAt     string `json:"start_at"`
		EndAt       string `json:"end_at"`
		Description string `json:"description"`
	}

	blocks := make([]routineBlockDTO, len(h.config.RoutineBlocks))
	for i, rb := range h.config.RoutineBlocks {
		blocks[i] = routineBlockDTO{
			Weekday:   int(rb.Weekday),
			StartTime: rb.StartTime,
			EndTime:   rb.EndTime,
			Label:     rb.Label,
			Tag:       rb.ProjectTag,
		}
	}

	evDTOs := make([]calEventDTO, len(calEvents))
	for i, e := range calEvents {
		evDTOs[i] = calEventDTO{
			ExternalID:  e.ExternalID,
			Source:      e.Source,
			Title:       e.Title,
			StartAt:     e.StartAt.Format(time.RFC3339),
			EndAt:       e.EndAt.Format(time.RFC3339),
			Description: e.Description,
		}
	}

	payload := struct {
		Type           string         `json:"type"`
		Tasks          []store.Task   `json:"tasks"`
		CalendarEvents []calEventDTO  `json:"calendar_events"`
		RoutineBlocks  []routineBlockDTO `json:"routine_blocks"`
		ServerTime     string         `json:"server_time"`
	}{
		Type:           "state",
		Tasks:          tasks,
		CalendarEvents: evDTOs,
		RoutineBlocks:  blocks,
		ServerTime:     time.Now().Format(time.RFC3339),
	}

	return json.Marshal(payload)
}

// Serve starts the HTTP server and blocks until ctx is cancelled.
func (h *Hub) Serve(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", h.handleWS)
	mux.HandleFunc("/favicon.svg", handleFavicon)
	// After the GCal OAuth callback is handled by the temporary auth server,
	// the browser still holds this URL. Redirect any repeat requests to root.
	mux.HandleFunc("/api/auth/callback/google", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	mux.HandleFunc("/", handleIndex)

	srv := &http.Server{
		Addr:        addr,
		Handler:     mux,
		ReadTimeout: 15 * time.Second,
	}

	// Shut down cleanly when ctx is cancelled.
	go func() { //nolint:gosec
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", addr, err)
	}

	slog.Info("nullcal web UI", "url", "http://"+addr)

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serving: %w", err)
	}
	return nil
}
