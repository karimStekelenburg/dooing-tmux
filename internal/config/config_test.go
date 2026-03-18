package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := Defaults()

	if cfg.Window.Width != 55 {
		t.Errorf("expected window width 55, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height != 20 {
		t.Errorf("expected window height 20, got %d", cfg.Window.Height)
	}
	if cfg.HourScoreValue != 0.125 {
		t.Errorf("expected HourScoreValue 0.125, got %f", cfg.HourScoreValue)
	}
	if len(cfg.Priorities) != 2 {
		t.Errorf("expected 2 default priorities, got %d", len(cfg.Priorities))
	}
	if cfg.IndentSize != 2 {
		t.Errorf("expected IndentSize 2, got %d", cfg.IndentSize)
	}
}

func TestLoadMissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.toml")
	if err != nil {
		t.Fatalf("expected no error for missing config, got: %v", err)
	}
	// Should return defaults
	if cfg.Window.Width != 55 {
		t.Errorf("expected default window width")
	}
}

func TestLoadTOMLOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")

	toml := `
[window]
width = 80
height = 30
`
	if err := os.WriteFile(path, []byte(toml), 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if cfg.Window.Width != 80 {
		t.Errorf("expected overridden width 80, got %d", cfg.Window.Width)
	}
	if cfg.Window.Height != 30 {
		t.Errorf("expected overridden height 30, got %d", cfg.Window.Height)
	}
	// Non-overridden defaults should remain
	if cfg.HourScoreValue != 0.125 {
		t.Errorf("expected default HourScoreValue to remain")
	}
}
