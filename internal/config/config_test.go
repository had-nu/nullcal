package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadFromPath_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
adapter: gcal
routine_blocks:
  - weekday: monday
    start_time: "09:00"
    end_time: "11:00"
    label: "Vexil -- feature dev"
    project_tag: vexil
  - weekday: wednesday
    start_time: "14:00"
    end_time: "16:00"
    label: "Wardex -- triage"
    project_tag: wardex
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath: %v", err)
	}

	if cfg.Adapter != "gcal" {
		t.Errorf("adapter = %q, want %q", cfg.Adapter, "gcal")
	}

	if len(cfg.RoutineBlocks) != 2 {
		t.Fatalf("routine_blocks count = %d, want 2", len(cfg.RoutineBlocks))
	}

	rb := cfg.RoutineBlocks[0]
	if rb.Weekday != time.Monday {
		t.Errorf("weekday = %v, want Monday", rb.Weekday)
	}
	if rb.StartTime != "09:00" {
		t.Errorf("start_time = %q, want %q", rb.StartTime, "09:00")
	}
	if rb.ProjectTag != "vexil" {
		t.Errorf("project_tag = %q, want %q", rb.ProjectTag, "vexil")
	}
}

func TestLoadFromPath_InvalidWeekday(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
routine_blocks:
  - weekday: funday
    start_time: "09:00"
    end_time: "11:00"
    label: "Bad"
    project_tag: test
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("expected error for invalid weekday, got nil")
	}
}

func TestLoadFromPath_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte(":::invalid"), 0o600); err != nil {
		t.Fatalf("writing test config: %v", err)
	}

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadFromPath_FileNotFound(t *testing.T) {
	_, err := LoadFromPath("/nonexistent/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_DefaultsWhenNoFile(t *testing.T) {
	// Point XDG to a temp dir with no config file.
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	t.Setenv("XDG_DATA_HOME", dir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Adapter != "" {
		t.Errorf("adapter = %q, want empty default", cfg.Adapter)
	}
	if len(cfg.RoutineBlocks) != 0 {
		t.Errorf("routine_blocks = %d, want 0", len(cfg.RoutineBlocks))
	}
}

func TestConfig_DBPath(t *testing.T) {
	cfg := &Config{DataDir: "/tmp/test-data"}
	want := "/tmp/test-data/nullcal.db"
	if got := cfg.DBPath(); got != want {
		t.Errorf("DBPath() = %q, want %q", got, want)
	}
}
