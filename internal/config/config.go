// Package config provides runtime configuration loading for nullcal.
//
// Configuration is loaded from $XDG_CONFIG_HOME/nullcal/config.yaml.
// If the file does not exist, sensible defaults are used.
package config // import "github.com/had-nu/nullcal/internal/config"

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the complete runtime configuration.
type Config struct {
	// DataDir is the resolved path for persistent data (SQLite DB).
	DataDir string `yaml:"-"`

	// ConfigDir is the resolved path for configuration files.
	ConfigDir string `yaml:"-"`

	// RoutineBlocks defines fixed weekly time slots tied to projects.
	RoutineBlocks []RoutineBlock `yaml:"routine_blocks"`

	// Adapter selects the calendar sync backend. Empty means no sync.
	Adapter string `yaml:"adapter"` // "gcal", "caldav", ""
}

// RoutineBlock is a fixed weekly time slot tied to a project.
// Defined in config.yaml, not in the database.
type RoutineBlock struct {
	Weekday    time.Weekday `yaml:"-"`
	WeekdayStr string       `yaml:"weekday"` // "monday", "tuesday", etc.
	StartTime  string       `yaml:"start_time"`
	EndTime    string       `yaml:"end_time"`
	Label      string       `yaml:"label"`
	ProjectTag string       `yaml:"project_tag"`
}

// Load reads configuration from the XDG config directory.
// If the config file does not exist, it returns defaults.
func Load() (*Config, error) {
	cfg := defaultConfig()

	dirs, err := resolveDirs()
	if err != nil {
		return nil, fmt.Errorf("resolving directories: %w", err)
	}
	cfg.ConfigDir = dirs.config
	cfg.DataDir = dirs.data

	path := filepath.Join(cfg.ConfigDir, "config.yaml")
	data, err := os.ReadFile(path) //nolint:gosec // config path is resolved from XDG, not user input
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.resolveWeekdays(); err != nil {
		return nil, fmt.Errorf("resolving weekdays: %w", err)
	}

	return cfg, nil
}

// LoadFromPath reads configuration from a specific file path.
// Used for testing and custom config locations.
func LoadFromPath(path string) (*Config, error) {
	cfg := defaultConfig()

	data, err := os.ReadFile(path) //nolint:gosec // caller provides the path explicitly
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.resolveWeekdays(); err != nil {
		return nil, fmt.Errorf("resolving weekdays: %w", err)
	}

	return cfg, nil
}

// DBPath returns the full path to the SQLite database file.
func (c *Config) DBPath() string {
	return filepath.Join(c.DataDir, "nullcal.db")
}

func defaultConfig() *Config {
	return &Config{
		RoutineBlocks: nil,
		Adapter:       "",
	}
}

type xdgDirs struct {
	config string
	data   string
}

func resolveDirs() (*xdgDirs, error) {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		configHome = filepath.Join(home, ".config")
	}

	dataHome := os.Getenv("XDG_DATA_HOME")
	if dataHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		dataHome = filepath.Join(home, ".local", "share")
	}

	dirs := &xdgDirs{
		config: filepath.Join(configHome, "nullcal"),
		data:   filepath.Join(dataHome, "nullcal"),
	}

	return dirs, nil
}

// resolveWeekdays converts the string weekday names to time.Weekday values.
func (c *Config) resolveWeekdays() error {
	weekdays := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
	}

	for i := range c.RoutineBlocks {
		wd, ok := weekdays[c.RoutineBlocks[i].WeekdayStr]
		if !ok {
			return fmt.Errorf("invalid weekday %q in routine block %q",
				c.RoutineBlocks[i].WeekdayStr, c.RoutineBlocks[i].Label)
		}
		c.RoutineBlocks[i].Weekday = wd
	}

	return nil
}
