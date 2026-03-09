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

// ListCalendarEvents returns all stored calendar events ordered by start time.
func (s *Store) ListCalendarEvents() ([]CalendarEvent, error) {
	const q = `SELECT external_id, source, title, start_at, end_at, description, synced_at
               FROM calendar_events ORDER BY start_at ASC`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("querying calendar events: %w", err)
	}
	defer rows.Close()

	var events []CalendarEvent
	for rows.Next() {
		var e CalendarEvent
		var startAt, endAt, syncedAt string
		if err := rows.Scan(
			&e.ExternalID, &e.Source, &e.Title,
			&startAt, &endAt, &e.Description, &syncedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning calendar event: %w", err)
		}
		// Parse timestamps stored as RFC3339.
		if t, err := time.Parse(time.RFC3339, startAt); err == nil {
			e.StartAt = t
		}
		if t, err := time.Parse(time.RFC3339, endAt); err == nil {
			e.EndAt = t
		}
		if t, err := time.Parse(time.RFC3339, syncedAt); err == nil {
			e.SyncedAt = t
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

