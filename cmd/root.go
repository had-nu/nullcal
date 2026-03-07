// Package cmd provides the CLI commands for nullcal.
package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/had-nu/nullcal/internal/config"
	"github.com/had-nu/nullcal/internal/store"
	"github.com/had-nu/nullcal/internal/tui"
)

const version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:   "nullcal",
	Short: "TUI-native calendar and task manager.",
	Long: `nullcal -- Your attention is the interface. Everything else is noise.

A local-first TUI calendar and task manager with week view and kanban board.
Google Calendar sync is an optional, replaceable backend.`,
	RunE: runTUI,
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
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("nullcal v%s\n", version)
	},
}

func runTUI(_ *cobra.Command, _ []string) error {
	// Load configuration.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Ensure data directory exists.
	if err := os.MkdirAll(cfg.DataDir, 0o750); err != nil {
		return fmt.Errorf("creating data directory: %w", err)
	}

	// Open database and run migrations.
	db, err := store.New(cfg.DBPath())
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	// Start TUI.
	m := tui.New(db, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}
