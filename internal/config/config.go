package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Priority defines a named priority with a weight for scoring.
type Priority struct {
	Name   string `toml:"name"`
	Weight int    `toml:"weight"`
}

// PriorityGroup defines a named group with member priorities and a color.
type PriorityGroup struct {
	Members []string `toml:"members"`
	Color   string   `toml:"color"`
}

// WindowConfig controls the main window appearance.
type WindowConfig struct {
	Width    int    `toml:"width"`
	Height   int    `toml:"height"`
	Border   string `toml:"border"`
	Position string `toml:"position"`
}

// CalendarConfig controls calendar display options.
type CalendarConfig struct {
	Language string `toml:"language"`
	StartDay string `toml:"start_day"` // "sunday" or "monday"
}

// DueNotificationsConfig controls due date notifications.
type DueNotificationsConfig struct {
	Enabled   bool `toml:"enabled"`
	OnStartup bool `toml:"on_startup"`
}

// Config is the root configuration structure.
type Config struct {
	Priorities          []Priority                `toml:"priorities"`
	PriorityGroups      map[string]PriorityGroup  `toml:"priority_groups"`
	HourScoreValue      float64                   `toml:"hour_score_value"`
	Window              WindowConfig              `toml:"window"`
	Calendar            CalendarConfig            `toml:"calendar"`
	DueNotifications    DueNotificationsConfig    `toml:"due_notifications"`
	DoneSortByCompleted bool                      `toml:"done_sort_by_completed_time"`
	MoveCompletedToEnd  bool                      `toml:"move_completed_to_end"`
	QuickKeys           bool                      `toml:"quick_keys"`
	IndentSize          int                       `toml:"indent_size"`
	ProjectFile         string                    `toml:"project_file"`
	AutoGitignore       string                    `toml:"auto_gitignore"` // "true", "false", "prompt"
	OnMissingProject    string                    `toml:"on_missing_project"` // "prompt", "auto_create"
}

// Defaults returns a Config with sensible defaults matching the original plugin.
func Defaults() Config {
	return Config{
		Priorities: []Priority{
			{Name: "important", Weight: 4},
			{Name: "urgent", Weight: 2},
		},
		PriorityGroups: map[string]PriorityGroup{
			"high": {
				Members: []string{"important", "urgent"},
				Color:   "#ff0000",
			},
			"medium": {
				Members: []string{"important"},
				Color:   "#ffff00",
			},
			"low": {
				Members: []string{"urgent"},
				Color:   "#0000ff",
			},
		},
		HourScoreValue: 0.125,
		Window: WindowConfig{
			Width:    55,
			Height:   20,
			Border:   "rounded",
			Position: "center",
		},
		Calendar: CalendarConfig{
			Language: "en",
			StartDay: "sunday",
		},
		DueNotifications: DueNotificationsConfig{
			Enabled:   true,
			OnStartup: true,
		},
		DoneSortByCompleted: true,
		MoveCompletedToEnd:  true,
		QuickKeys:           true,
		IndentSize:          2,
		ProjectFile:         "dooing.json",
		AutoGitignore:       "prompt",
		OnMissingProject:    "prompt",
	}
}

// DefaultConfigPath returns the XDG config path for the config file.
func DefaultConfigPath() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "dooing", "config.toml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "config.toml"
	}
	return filepath.Join(home, ".config", "dooing", "config.toml")
}

// Load reads the config from path, falling back to defaults for any missing fields.
// If the file does not exist, full defaults are returned without error.
func Load(path string) (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}

	// Deep-merge: decode into the defaults struct so missing keys retain defaults.
	if _, err := toml.Decode(string(data), &cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
