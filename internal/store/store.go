package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// DefaultPath returns the default path for the global todos file, following XDG.
func DefaultPath() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return filepath.Join(xdg, "dooing", "todos.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "todos.json"
	}
	return filepath.Join(home, ".local", "share", "dooing", "todos.json")
}

// Store handles loading and saving todos to disk.
type Store struct{}

// New returns a new Store.
func New() *Store {
	return &Store{}
}

// Load reads todos from path, creating the file with an empty list if it doesn't exist.
// It also runs migrations to backfill any missing fields.
func (s *Store) Load(path string) ([]*model.Todo, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []*model.Todo{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading todos file: %w", err)
	}

	var todos []*model.Todo
	if err := json.Unmarshal(data, &todos); err != nil {
		return nil, fmt.Errorf("parsing todos: %w", err)
	}

	migrate(todos)
	return todos, nil
}

// Save atomically writes todos to path using a temp file + rename.
func (s *Store) Save(path string, todos []*model.Todo) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating data dir: %w", err)
	}

	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding todos: %w", err)
	}

	// Write to a temp file in the same directory so rename is atomic.
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".dooing-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}

// migrate backfills any missing fields on legacy todos.
func migrate(todos []*model.Todo) {
	for _, t := range todos {
		// Backfill ID if missing (legacy todos had no ID).
		if t.ID == "" {
			nt := model.NewTodo(t.Text)
			t.ID = nt.ID
		}
		// ParentID, Depth, Notes already have zero values that are valid defaults.
	}
}
