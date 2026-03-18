package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	s := New()
	todos := []*model.Todo{
		model.NewTodo("buy milk #shopping"),
		model.NewTodo("write tests #work"),
	}

	if err := s.Save(path, todos); err != nil {
		t.Fatalf("save error: %v", err)
	}

	loaded, err := s.Load(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}

	if len(loaded) != 2 {
		t.Errorf("expected 2 todos, got %d", len(loaded))
	}
	if loaded[0].Text != "buy milk #shopping" {
		t.Errorf("unexpected text: %s", loaded[0].Text)
	}
}

func TestLoadMissingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nonexistent", "todos.json")

	s := New()
	todos, err := s.Load(path)
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected empty list, got %d todos", len(todos))
	}
}

func TestMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.json")

	// Write a legacy JSON file with no ID field.
	legacy := `[{"text":"legacy todo","done":false}]`
	if err := os.WriteFile(path, []byte(legacy), 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	s := New()
	todos, err := s.Load(path)
	if err != nil {
		t.Fatalf("load error: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(todos))
	}
	if todos[0].ID == "" {
		t.Error("expected ID to be backfilled on migration")
	}
}

func TestAtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")

	s := New()
	todos := []*model.Todo{model.NewTodo("atomic test")}

	if err := s.Save(path, todos); err != nil {
		t.Fatalf("save error: %v", err)
	}

	// Verify no leftover temp files
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name() != "todos.json" {
			t.Errorf("unexpected leftover file: %s", e.Name())
		}
	}
}
