// Package store defines the domain types used across nullcal.
package store

import "time"

// TaskStatus represents the current state of a task in the kanban workflow.
type TaskStatus string

// TaskStatusBacklog and related constants define the kanban workflow states.
const (
	TaskStatusTodo    TaskStatus = "todo"
	TaskStatusBacklog TaskStatus = "backlog"
	TaskStatusDoing   TaskStatus = "doing"
	TaskStatusDone    TaskStatus = "done"
)

// ValidTaskStatus reports whether s is a recognized task status.
func ValidTaskStatus(s TaskStatus) bool {
	switch s {
	case TaskStatusTodo, TaskStatusBacklog, TaskStatusDoing, TaskStatusDone:
		return true
	}
	return false
}

// Recurrence defines the repetition pattern for a task.
// Empty string means no recurrence.
type Recurrence string

// RecurrenceNone and related constants define repetition patterns.
const (
	RecurrenceNone    Recurrence = ""
	RecurrenceDaily   Recurrence = "daily"
	RecurrenceWeekly  Recurrence = "weekly"
	RecurrenceMonthly Recurrence = "monthly"
)

// Task is a local unit of work, independent of any external calendar.
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ProjectTag  string     `json:"project_tag"`
	Status      TaskStatus `json:"status"`
	DueAt       *time.Time `json:"due_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Recurrence  Recurrence `json:"recurrence"`
	CreatedAt   time.Time  `json:"created_at"`
	GCalEventID *string    `json:"gcal_event_id,omitempty"`
}

// CalendarEvent is the normalised representation of an external event.
// Populated by any CalendarAdapter implementation.
type CalendarEvent struct {
	ExternalID  string    `json:"external_id"`
	Source      string    `json:"source"`
	Title       string    `json:"title"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Description string    `json:"description"`
	SyncedAt    time.Time `json:"synced_at"`
}
