package ui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func makeTodo(text string) *model.Todo {
	t := model.NewTodo(text)
	return t
}

func TestExportTodos(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "export.json")

	todos := []*model.Todo{
		makeTodo("buy milk #shopping"),
		makeTodo("write tests #dev"),
	}

	if err := exportTodos(todos, path); err != nil {
		t.Fatalf("exportTodos: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading export file: %v", err)
	}

	var loaded []*model.Todo
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("parsing export file: %v", err)
	}

	if len(loaded) != len(todos) {
		t.Errorf("got %d todos, want %d", len(loaded), len(todos))
	}
	for i, td := range loaded {
		if td.Text != todos[i].Text {
			t.Errorf("todo[%d] text = %q, want %q", i, td.Text, todos[i].Text)
		}
	}
}

func TestImportTodos_Merge(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "import.json")

	incoming := []*model.Todo{
		makeTodo("new task #work"),
	}
	data, _ := json.Marshal(incoming)
	_ = os.WriteFile(path, data, 0o644)

	existing := []*model.Todo{
		makeTodo("existing task #personal"),
	}

	merged, err := importTodos(existing, path)
	if err != nil {
		t.Fatalf("importTodos: %v", err)
	}
	if len(merged) != 2 {
		t.Errorf("got %d todos, want 2", len(merged))
	}
}

func TestImportTodos_Deduplication(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "import.json")

	existing := []*model.Todo{makeTodo("shared task")}

	// Export existing to a file, then import it back — should deduplicate.
	data, _ := json.Marshal(existing)
	_ = os.WriteFile(path, data, 0o644)

	merged, err := importTodos(existing, path)
	if err != nil {
		t.Fatalf("importTodos: %v", err)
	}
	if len(merged) != 1 {
		t.Errorf("got %d todos after dedup, want 1", len(merged))
	}
}

func TestImportTodos_FileNotFound(t *testing.T) {
	_, err := importTodos(nil, "/nonexistent/path/todos.json")
	if err == nil {
		t.Error("expected error for missing file, got nil")
	}
}

func TestImportTodos_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	_ = os.WriteFile(path, []byte("{not json}"), 0o644)

	_, err := importTodos(nil, path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestDeduplicateTodos(t *testing.T) {
	a := makeTodo("hello")
	b := makeTodo("hello") // same text, different ID/timestamp

	// Manually set same non-ID fields to trigger hash collision.
	b.Text = a.Text
	b.Done = a.Done
	b.InProgress = a.InProgress
	b.Category = a.Category

	result := deduplicateTodos([]*model.Todo{a, b})
	if len(result) != 1 {
		t.Errorf("got %d todos, want 1", len(result))
	}
	if result[0].ID != a.ID {
		t.Error("expected first occurrence to be kept")
	}
}

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rt.json")

	original := []*model.Todo{
		makeTodo("task one #work"),
		makeTodo("task two #personal"),
	}
	original[0].Done = true
	original[1].Priorities = []string{"important"}

	if err := exportTodos(original, path); err != nil {
		t.Fatalf("export: %v", err)
	}

	merged, err := importTodos([]*model.Todo{}, path)
	if err != nil {
		t.Fatalf("import: %v", err)
	}

	if len(merged) != len(original) {
		t.Errorf("round-trip: got %d, want %d", len(merged), len(original))
	}
	for i, td := range merged {
		if td.Text != original[i].Text {
			t.Errorf("round-trip text[%d]: got %q, want %q", i, td.Text, original[i].Text)
		}
		if td.Done != original[i].Done {
			t.Errorf("round-trip done[%d]: got %v, want %v", i, td.Done, original[i].Done)
		}
	}
}
