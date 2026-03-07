package store

import (
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	s, err := New(":memory:")
	if err != nil {
		t.Fatalf("creating test store: %v", err)
	}

	if err := s.Migrate(); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestMigrate_Idempotent(t *testing.T) {
	s := newTestStore(t)

	// Running migrate again should not fail.
	if err := s.Migrate(); err != nil {
		t.Fatalf("second migration run: %v", err)
	}
}

func TestCreateTask(t *testing.T) {
	s := newTestStore(t)

	task := &Task{
		Title:      "implement VEX reachability",
		ProjectTag: "wardex",
	}

	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if task.ID == "" {
		t.Error("expected auto-generated ID, got empty")
	}
	if task.Status != TaskStatusBacklog {
		t.Errorf("status = %q, want %q", task.Status, TaskStatusBacklog)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestCreateTask_InvalidStatus(t *testing.T) {
	s := newTestStore(t)

	task := &Task{
		Title:  "bad status",
		Status: "invalid",
	}

	if err := s.CreateTask(task); err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestGetTask(t *testing.T) {
	s := newTestStore(t)

	due := time.Date(2026, 3, 10, 14, 0, 0, 0, time.UTC)
	task := &Task{
		Title:       "write tests",
		Description: "store layer unit tests",
		ProjectTag:  "nullcal",
		DueAt:       &due,
		Recurrence:  RecurrenceWeekly,
	}

	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	got, err := s.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}

	if got.Title != task.Title {
		t.Errorf("title = %q, want %q", got.Title, task.Title)
	}
	if got.Description != task.Description {
		t.Errorf("description = %q, want %q", got.Description, task.Description)
	}
	if got.ProjectTag != task.ProjectTag {
		t.Errorf("project_tag = %q, want %q", got.ProjectTag, task.ProjectTag)
	}
	if got.DueAt == nil || !got.DueAt.Equal(due) {
		t.Errorf("due_at = %v, want %v", got.DueAt, due)
	}
	if got.Recurrence != RecurrenceWeekly {
		t.Errorf("recurrence = %q, want %q", got.Recurrence, RecurrenceWeekly)
	}
}

func TestGetTask_NotFound(t *testing.T) {
	s := newTestStore(t)

	_, err := s.GetTask("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent task, got nil")
	}
}

func TestListTasks(t *testing.T) {
	s := newTestStore(t)

	for i := 0; i < 3; i++ {
		task := &Task{Title: "task"}
		if err := s.CreateTask(task); err != nil {
			t.Fatalf("CreateTask %d: %v", i, err)
		}
	}

	tasks, err := s.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}

	if len(tasks) != 3 {
		t.Errorf("count = %d, want 3", len(tasks))
	}
}

func TestListTasksByStatus(t *testing.T) {
	s := newTestStore(t)

	backlog := &Task{Title: "backlog task", Status: TaskStatusBacklog}
	doing := &Task{Title: "doing task", Status: TaskStatusDoing}

	for _, task := range []*Task{backlog, doing} {
		if err := s.CreateTask(task); err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
	}

	tasks, err := s.ListTasksByStatus(TaskStatusBacklog)
	if err != nil {
		t.Fatalf("ListTasksByStatus: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("count = %d, want 1", len(tasks))
	}
	if tasks[0].Title != "backlog task" {
		t.Errorf("title = %q, want %q", tasks[0].Title, "backlog task")
	}
}

func TestListTasksByDate(t *testing.T) {
	s := newTestStore(t)

	today := time.Date(2026, 3, 10, 10, 0, 0, 0, time.UTC)
	tomorrow := time.Date(2026, 3, 11, 10, 0, 0, 0, time.UTC)

	todayTask := &Task{Title: "today", DueAt: &today}
	tomorrowTask := &Task{Title: "tomorrow", DueAt: &tomorrow}

	for _, task := range []*Task{todayTask, tomorrowTask} {
		if err := s.CreateTask(task); err != nil {
			t.Fatalf("CreateTask: %v", err)
		}
	}

	tasks, err := s.ListTasksByDate(today)
	if err != nil {
		t.Fatalf("ListTasksByDate: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("count = %d, want 1", len(tasks))
	}
	if tasks[0].Title != "today" {
		t.Errorf("title = %q, want %q", tasks[0].Title, "today")
	}
}

func TestUpdateTask(t *testing.T) {
	s := newTestStore(t)

	task := &Task{Title: "original"}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	task.Title = "updated"
	task.ProjectTag = "vexil"
	if err := s.UpdateTask(task); err != nil {
		t.Fatalf("UpdateTask: %v", err)
	}

	got, err := s.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Title != "updated" {
		t.Errorf("title = %q, want %q", got.Title, "updated")
	}
	if got.ProjectTag != "vexil" {
		t.Errorf("project_tag = %q, want %q", got.ProjectTag, "vexil")
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	s := newTestStore(t)

	task := &Task{ID: "nonexistent", Title: "nope", Status: TaskStatusBacklog}
	if err := s.UpdateTask(task); err == nil {
		t.Fatal("expected error for nonexistent task, got nil")
	}
}

func TestDeleteTask(t *testing.T) {
	s := newTestStore(t)

	task := &Task{Title: "to delete"}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := s.DeleteTask(task.ID); err != nil {
		t.Fatalf("DeleteTask: %v", err)
	}

	_, err := s.GetTask(task.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	s := newTestStore(t)

	if err := s.DeleteTask("nonexistent"); err == nil {
		t.Fatal("expected error for nonexistent task, got nil")
	}
}

func TestSetTaskStatus_BacklogToDoing(t *testing.T) {
	s := newTestStore(t)

	task := &Task{Title: "transition test"}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := s.SetTaskStatus(task.ID, TaskStatusDoing); err != nil {
		t.Fatalf("SetTaskStatus: %v", err)
	}

	got, err := s.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Status != TaskStatusDoing {
		t.Errorf("status = %q, want %q", got.Status, TaskStatusDoing)
	}
	if got.CompletedAt != nil {
		t.Error("expected nil CompletedAt for doing status")
	}
}

func TestSetTaskStatus_DoingToDone(t *testing.T) {
	s := newTestStore(t)

	task := &Task{Title: "complete me", Status: TaskStatusDoing}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := s.SetTaskStatus(task.ID, TaskStatusDone); err != nil {
		t.Fatalf("SetTaskStatus: %v", err)
	}

	got, err := s.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask: %v", err)
	}
	if got.Status != TaskStatusDone {
		t.Errorf("status = %q, want %q", got.Status, TaskStatusDone)
	}
	if got.CompletedAt == nil {
		t.Fatal("expected CompletedAt to be set for done status")
	}
}

func TestSetTaskStatus_InvalidStatus(t *testing.T) {
	s := newTestStore(t)

	task := &Task{Title: "bad transition"}
	if err := s.CreateTask(task); err != nil {
		t.Fatalf("CreateTask: %v", err)
	}

	if err := s.SetTaskStatus(task.ID, "invalid"); err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
}

func TestSetTaskStatus_NotFound(t *testing.T) {
	s := newTestStore(t)

	if err := s.SetTaskStatus("nonexistent", TaskStatusDone); err == nil {
		t.Fatal("expected error for nonexistent task, got nil")
	}
}

func TestEmptyDatabase(t *testing.T) {
	s := newTestStore(t)

	tasks, err := s.ListTasks()
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected empty list, got %d tasks", len(tasks))
	}
}
