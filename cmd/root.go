// Package cmd provides the CLI commands for nullcal.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
	"github.com/had-nu/nullcal/internal/sync/gcal"
	"github.com/had-nu/nullcal/internal/web"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "nullcal",
	Short: "Local-first calendar and task manager.",
	Long: `nullcal -- Your attention is the interface. Everything else is noise.

A local web application providing a week view and kanban board.
Google Calendar sync is an optional, replaceable backend.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Default behavior: just run the server
		return runWeb(serveAddr)
	},
}

var serveAddr string
var gcalSync bool

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Launch the browser UI over WebSocket",
	Long: `Start a local HTTP server and open the nullcal browser dashboard.

The browser connects to the server via WebSocket. Every action (create, edit,
delete, move status) is sent to the server, persisted in SQLite, and broadcast
back to all connected tabs in real time.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runWeb(serveAddr)
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(versionCmd)

	serveCmd.Flags().StringVarP(&serveAddr, "addr", "a", "localhost:7331", "address to listen on")
	serveCmd.Flags().BoolVar(&gcalSync, "gcal-sync", false, "sync events from Google Calendar on startup")
	rootCmd.Flags().BoolVar(&gcalSync, "gcal-sync", false, "sync events from Google Calendar on startup")
	rootCmd.AddCommand(serveCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("nullcal v%s\n", version)
	},
}

func runWeb(addr string) error {
	// Load .env before anything else so GOOGLE_* vars are available.
	if err := loadEnv(".env"); err != nil {
		fmt.Fprintf(os.Stderr, "note: .env not loaded: %v\n", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	db, err := store.New(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Shut down on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		time.Sleep(300 * time.Millisecond) // Give the server a moment to bind
		url := "http://" + addr
		if addr[0] == ':' {
			url = "http://localhost" + addr
		}
		openBrowser(url)
	}()

	hub := web.NewHub(db, cfg)

	// Optional GCal sync: fetch events and store them before serving.
	if gcalSync {
		fmt.Println("[gcal] Starting Google Calendar sync...")
		syncCtx, syncCancel := context.WithTimeout(ctx, 60*time.Second)
		defer syncCancel()
		adapter, err := gcal.New(syncCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[gcal] auth failed: %v\n", err)
		} else {
			now := time.Now()
			events, err := adapter.ListEvents(syncCtx, now.AddDate(0, -1, 0), now.AddDate(0, 3, 0))
			if err != nil {
				fmt.Fprintf(os.Stderr, "[gcal] fetch failed: %v\n", err)
			} else {
				if err := db.UpsertCalendarEvents(events); err != nil {
					fmt.Fprintf(os.Stderr, "[gcal] store failed: %v\n", err)
				} else {
					fmt.Printf("[gcal] synced %d events.\n", len(events))
				}
			}
		}
	}

	return hub.Serve(ctx, addr)
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start() //nolint:gosec
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start() //nolint:gosec
	case "darwin":
		err = exec.Command("open", url).Start() //nolint:gosec
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser automatically: %v\nOpen %s manually.\n", err, url)
	}
}

// loadEnv reads a .env file and sets missing environment variables.
// Lines starting with # and empty lines are ignored.
// Silently returns nil if the file does not exist.
func loadEnv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // no .env is fine
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		// Only set if not already provided by the environment.
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

