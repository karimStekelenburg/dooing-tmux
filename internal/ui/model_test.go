package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func tempModel(t *testing.T) Model {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "todos.json")
	if err := os.WriteFile(path, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := NewModel(false)
	m.storePath = path
	m.todos = m.todos[:0]
	return m
}

func sendKey(m Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(Model)
}

func sendSpecialKey(m Model, k tea.KeyType) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: k})
	return updated.(Model)
}

func typeText(m Model, text string) Model {
	for _, ch := range text {
		m = sendKey(m, string(ch))
	}
	return m
}

func TestNewModel(t *testing.T) {
	m := NewModel(false)
	if m.projectMode {
		t.Error("expected projectMode=false")
	}
	m2 := NewModel(true)
	if !m2.projectMode {
		t.Error("expected projectMode=true")
	}
}

func TestInitReturnsNil(t *testing.T) {
	m := NewModel(false)
	cmd := m.Init()
	if cmd != nil {
		t.Error("expected Init() to return nil cmd")
	}
}

func TestQuitKey(t *testing.T) {
	m := NewModel(false)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected a quit command from q key")
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := NewModel(false)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	um := updated.(Model)
	if um.width != 80 || um.height != 24 {
		t.Errorf("expected 80x24, got %dx%d", um.width, um.height)
	}
}

func TestViewContainsTitle(t *testing.T) {
	m := NewModel(false)
	view := m.View()
	if len(view) == 0 {
		t.Error("expected non-empty view")
	}
}

func TestCreateTodo(t *testing.T) {
	m := tempModel(t)

	m = sendKey(m, "i")
	if m.inputMode != inputModeCreate {
		t.Fatal("expected inputModeCreate after pressing i")
	}

	m = typeText(m, "Buy groceries #shopping")
	m = sendSpecialKey(m, tea.KeyEnter)

	if m.inputMode != inputModeNone {
		t.Fatal("expected inputModeNone after Enter")
	}
	if len(m.todos) != 1 {
		t.Fatalf("expected 1 todo, got %d", len(m.todos))
	}
	if m.todos[0].Text != "Buy groceries #shopping" {
		t.Errorf("unexpected text: %q", m.todos[0].Text)
	}
	if m.todos[0].Category != "shopping" {
		t.Errorf("expected category 'shopping', got %q", m.todos[0].Category)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}
}

func TestCreateTodoEmpty(t *testing.T) {
	m := tempModel(t)
	m = sendKey(m, "i")
	m = sendSpecialKey(m, tea.KeyEnter)

	if len(m.todos) != 0 {
		t.Fatalf("empty todo should not be created, got %d todos", len(m.todos))
	}
}

func TestCreateTodoEscCancels(t *testing.T) {
	m := tempModel(t)
	m = sendKey(m, "i")
	m = typeText(m, "Some text")
	m = sendSpecialKey(m, tea.KeyEsc)

	if m.inputMode != inputModeNone {
		t.Fatal("expected inputModeNone after Esc")
	}
	if len(m.todos) != 0 {
		t.Fatal("Esc should cancel without creating a todo")
	}
}

func TestEditTodo(t *testing.T) {
	m := tempModel(t)

	// Create a todo first.
	m = sendKey(m, "i")
	m = typeText(m, "Old text #work")
	m = sendSpecialKey(m, tea.KeyEnter)

	originalID := m.todos[0].ID

	// Edit it.
	m = sendKey(m, "e")
	if m.inputMode != inputModeEdit {
		t.Fatal("expected inputModeEdit after pressing e")
	}
	if m.editingID != originalID {
		t.Errorf("editingID mismatch")
	}

	// Overwrite value directly to simulate typing new text.
	m.ti.SetValue("New text #personal")
	m = sendSpecialKey(m, tea.KeyEnter)

	if m.inputMode != inputModeNone {
		t.Fatal("expected inputModeNone after edit submit")
	}
	if m.todos[0].Text != "New text #personal" {
		t.Errorf("unexpected text after edit: %q", m.todos[0].Text)
	}
	if m.todos[0].Category != "personal" {
		t.Errorf("expected category 'personal', got %q", m.todos[0].Category)
	}
}

func TestNavigation(t *testing.T) {
	m := tempModel(t)

	for _, text := range []string{"First", "Second"} {
		m = sendKey(m, "i")
		m = typeText(m, text)
		m = sendSpecialKey(m, tea.KeyEnter)
	}

	if m.cursor != 1 {
		t.Errorf("cursor should be at last created (1), got %d", m.cursor)
	}

	m = sendKey(m, "k")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after k, got %d", m.cursor)
	}
	m = sendKey(m, "k") // should clamp
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}

	m = sendKey(m, "j")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.cursor)
	}
	m = sendKey(m, "j") // should clamp
	if m.cursor != 1 {
		t.Errorf("cursor should stay at 1, got %d", m.cursor)
	}
}

func TestViewRenders(t *testing.T) {
	m := tempModel(t)
	v := m.View()
	if v == "" {
		t.Fatal("View should not return empty string")
	}
}
