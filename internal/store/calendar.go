package store

import (
	"fmt"
	"time"
)

// UpsertCalendarEvents inserts or replaces calendar events from an external source.
// Events are keyed by ExternalID + Source; existing entries are overwritten.
func (s *Store) UpsertCalendarEvents(events []CalendarEvent) error {
	if len(events) == 0 {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	const upsertSQL = `
		INSERT INTO calendar_events
			(external_id, source, title, start_at, end_at, description, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(external_id) DO UPDATE SET
			source      = excluded.source,
			title       = excluded.title,
			start_at    = excluded.start_at,
			end_at      = excluded.end_at,
			description = excluded.description,
			synced_at   = excluded.synced_at`

	stmt, err := tx.Prepare(upsertSQL)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("preparing upsert: %w", err)
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.Exec(
			e.ExternalID,
			e.Source,
			e.Title,
			e.StartAt.Format(time.RFC3339),
			e.EndAt.Format(time.RFC3339),
			e.Description,
			time.Now().Format(time.RFC3339),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("upserting event %s: %w", e.ExternalID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing upsert: %w", err)
	}
	return nil
}
