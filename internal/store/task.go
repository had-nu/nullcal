package store

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateTask inserts a new task into the database.
// The task ID is generated automatically if empty.
func (s *Store) CreateTask(t *Task) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	if t.Status == "" {
		t.Status = TaskStatusBacklog
	}
	if !ValidTaskStatus(t.Status) {
		return fmt.Errorf("invalid task status: %q", t.Status)
	}

	now := time.Now().UTC()
	t.CreatedAt = now

	_, err := s.db.Exec(`
		INSERT INTO tasks (id, title, description, project_tag, status, due_at, completed_at, recurrence, created_at, gcal_event_id, pomodoros)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.Title, t.Description, t.ProjectTag, t.Status,
		nullTime(t.DueAt), nullTime(t.CompletedAt), string(t.Recurrence), t.CreatedAt, nullString(t.GCalEventID), t.Pomodoros,
	)
	if err != nil {
		return fmt.Errorf("inserting task: %w", err)
	}

	return nil
}

// GetTask retrieves a single task by ID.
func (s *Store) GetTask(id string) (*Task, error) {
	row := s.db.QueryRow(`
		SELECT id, title, description, project_tag, status, due_at, completed_at, recurrence, created_at, gcal_event_id, pomodoros
		FROM tasks WHERE id = ?`, id)

	t, err := scanTask(row)
	if err != nil {
		return nil, fmt.Errorf("getting task %s: %w", id, err)
	}
	return t, nil
}

// ListTasks returns all tasks ordered by creation date descending.
func (s *Store) ListTasks() ([]Task, error) {
	rows, err := s.db.Query(`
		SELECT id, title, description, project_tag, status, due_at, completed_at, recurrence, created_at, gcal_event_id, pomodoros
		FROM tasks ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanTasks(rows)
}

// ListTasksByStatus returns all tasks with the given status.
func (s *Store) ListTasksByStatus(status TaskStatus) ([]Task, error) {
	if !ValidTaskStatus(status) {
		return nil, fmt.Errorf("invalid task status: %q", status)
	}

	rows, err := s.db.Query(`
		SELECT id, title, description, project_tag, status, due_at, completed_at, recurrence, created_at, gcal_event_id, pomodoros
		FROM tasks WHERE status = ? ORDER BY created_at DESC`, status)
	if err != nil {
		return nil, fmt.Errorf("listing tasks by status %s: %w", status, err)
	}
	defer func() { _ = rows.Close() }()

	return scanTasks(rows)
}

// ListTasksByDate returns all tasks with a due date on the given day.
func (s *Store) ListTasksByDate(date time.Time) ([]Task, error) {
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	rows, err := s.db.Query(`
		SELECT id, title, description, project_tag, status, due_at, completed_at, recurrence, created_at, gcal_event_id, pomodoros
		FROM tasks WHERE due_at >= ? AND due_at < ? ORDER BY due_at ASC`, start, end)
	if err != nil {
		return nil, fmt.Errorf("listing tasks by date: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanTasks(rows)
}

// UpdateTask updates an existing task.
func (s *Store) UpdateTask(t *Task) error {
	if !ValidTaskStatus(t.Status) {
		return fmt.Errorf("invalid task status: %q", t.Status)
	}

	result, err := s.db.Exec(`
		UPDATE tasks SET title = ?, description = ?, project_tag = ?, status = ?,
		due_at = ?, completed_at = ?, recurrence = ?, gcal_event_id = ?, pomodoros = ?
		WHERE id = ?`,
		t.Title, t.Description, t.ProjectTag, t.Status,
		nullTime(t.DueAt), nullTime(t.CompletedAt), string(t.Recurrence), nullString(t.GCalEventID), t.Pomodoros, t.ID,
	)
	if err != nil {
		return fmt.Errorf("updating task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task %s not found", t.ID)
	}

	return nil
}

// DeleteTask removes a task by ID.
func (s *Store) DeleteTask(id string) error {
	result, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking delete result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	return nil
}

// SetTaskStatus updates the status of a task and sets completed_at
// when transitioning to done. For recurring tasks, it also spawns the next occurrence.
func (s *Store) SetTaskStatus(id string, status TaskStatus) error {
	if !ValidTaskStatus(status) {
		return fmt.Errorf("invalid task status: %q", status)
	}

	// For recurrence spawning, we need the original task if marking as done
	var t *Task
	var err error
	if status == TaskStatusDone {
		t, err = s.GetTask(id)
		if err != nil {
			return err
		}
	}

	var completedAt *time.Time
	if status == TaskStatusDone {
		now := time.Now().UTC()
		completedAt = &now
	}

	result, err := s.db.Exec(
		"UPDATE tasks SET status = ?, completed_at = ? WHERE id = ?",
		status, nullTime(completedAt), id,
	)
	if err != nil {
		return fmt.Errorf("setting task status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking status update result: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("task %s not found", id)
	}

	// Spawn next recurrence if applicable and transitioning to done
	if status == TaskStatusDone && t != nil && t.Recurrence != RecurrenceNone && t.Recurrence != "" {
		nextDue := time.Now().UTC() // fallback
		if t.DueAt != nil {
			nextDue = *t.DueAt
		}
		
		switch t.Recurrence {
		case RecurrenceDaily:
			nextDue = nextDue.AddDate(0, 0, 1)
		case RecurrenceWeekly:
			nextDue = nextDue.AddDate(0, 0, 7)
		case RecurrenceMonthly:
			nextDue = nextDue.AddDate(0, 1, 0)
		}

		nextTask := &Task{
			Title:       t.Title,
			Description: t.Description,
			ProjectTag:  t.ProjectTag,
			Status:      TaskStatusTodo,
			DueAt:       &nextDue,
			Recurrence:  t.Recurrence,
			Pomodoros:   t.Pomodoros,
			// ID generated, CompletedAt nil, GCalEventID nil
		}
		if err := s.CreateTask(nextTask); err != nil {
			return fmt.Errorf("spawning recurring task: %w", err)
		}
	}

	return nil
}

// scanTask scans a single task from a database row.
func scanTask(row *sql.Row) (*Task, error) {
	var t Task
	var dueAt, completedAt sql.NullTime
	var recurrence, gcalEventID sql.NullString

	err := row.Scan(
		&t.ID, &t.Title, &t.Description, &t.ProjectTag, &t.Status,
		&dueAt, &completedAt, &recurrence, &t.CreatedAt, &gcalEventID, &t.Pomodoros,
	)
	if err != nil {
		return nil, err
	}

	if dueAt.Valid {
		t.DueAt = &dueAt.Time
	}
	if completedAt.Valid {
		t.CompletedAt = &completedAt.Time
	}
	if recurrence.Valid {
		t.Recurrence = Recurrence(recurrence.String)
	}
	if gcalEventID.Valid {
		t.GCalEventID = &gcalEventID.String
	}

	return &t, nil
}

// scanTasks scans multiple tasks from database rows.
func scanTasks(rows *sql.Rows) ([]Task, error) {
	var tasks []Task
	for rows.Next() {
		var t Task
		var dueAt, completedAt sql.NullTime
		var recurrence, gcalEventID sql.NullString

		err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.ProjectTag, &t.Status,
			&dueAt, &completedAt, &recurrence, &t.CreatedAt, &gcalEventID, &t.Pomodoros,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning task row: %w", err)
		}

		if dueAt.Valid {
			t.DueAt = &dueAt.Time
		}
		if completedAt.Valid {
			t.CompletedAt = &completedAt.Time
		}
		if recurrence.Valid {
			t.Recurrence = Recurrence(recurrence.String)
		}
		if gcalEventID.Valid {
			t.GCalEventID = &gcalEventID.String
		}

		tasks = append(tasks, t)
	}

	return tasks, rows.Err()
}

// nullTime converts a *time.Time to sql.NullTime for database operations.
func nullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func nullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}
