// Package store provides SQLite-backed persistence for nullcal.
//
// All data is stored locally in a single SQLite database with WAL mode
// enabled. Migrations are embedded and applied automatically on startup.
package store

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3" // SQLite driver registration
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Store wraps a SQLite database connection and provides typed access
// to nullcal's persistent data.
type Store struct {
	db *sql.DB
}

// New opens or creates a SQLite database at the given path.
// WAL mode and foreign keys are enabled on the connection.
// If dbPath is ":memory:", an in-memory database is used.
func New(dbPath string) (*Store, error) {
	if dbPath != ":memory:" {
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, fmt.Errorf("creating data directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the underlying database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// Migrate applies all pending SQL migrations in order.
// Migrations are embedded in the binary and named as NNN_description.sql.
func (s *Store) Migrate() error {
	// Create the migration tracking table if it does not exist.
	const createTable = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`
	if _, err := s.db.Exec(createTable); err != nil {
		return fmt.Errorf("creating schema_migrations table: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("reading migrations directory: %w", err)
	}

	// Sort by filename to guarantee order.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")

		// Check if already applied.
		var exists int
		err := s.db.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("checking migration %s: %w", version, err)
		}
		if exists > 0 {
			continue
		}

		// Read and execute the migration.
		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", version, err)
		}

		tx, err := s.db.Begin()
		if err != nil {
			return fmt.Errorf("beginning transaction for %s: %w", version, err)
		}

		if _, err := tx.Exec(string(data)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("executing migration %s: %w", version, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES (?)",
			version,
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("recording migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %s: %w", version, err)
		}
	}

	return nil
}
