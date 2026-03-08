// Package cmd provides the CLI commands for nullcal.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
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
