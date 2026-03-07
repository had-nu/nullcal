// Package store defines the domain types used across nullcal.
package store

import "time"

// TaskStatus represents the current state of a task in the kanban workflow.
type TaskStatus string

// TaskStatusBacklog and related constants define the kanban workflow states.
const (
	TaskStatusBacklog TaskStatus = "backlog"
	TaskStatusDoing   TaskStatus = "doing"
	TaskStatusDone    TaskStatus = "done"
)

// ValidTaskStatus reports whether s is a recognized task status.
func ValidTaskStatus(s TaskStatus) bool {
	switch s {
	case TaskStatusBacklog, TaskStatusDoing, TaskStatusDone:
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
	ID          string
	Title       string
	Description string
	ProjectTag  string
	Status      TaskStatus
	DueAt       *time.Time
	CompletedAt *time.Time
	Recurrence  Recurrence
	CreatedAt   time.Time
}

// CalendarEvent is the normalised representation of an external event.
// Populated by any CalendarAdapter implementation.
type CalendarEvent struct {
	ExternalID  string
	Source      string // "gcal", "caldav"
	Title       string
	StartAt     time.Time
	EndAt       time.Time
	Description string
	SyncedAt    time.Time
}
