package ui

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
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

// createTodo is a test helper that creates a todo via the UI flow.
func createTodo(m Model, text string) Model {
	m = sendKey(m, "i")
	m = typeText(m, text)
	return sendSpecialKey(m, tea.KeyEnter)
}

// navigateTo sets the cursor to the first todo whose Text equals text.
// It is used in tests to avoid brittle cursor-index assumptions after sorting.
func navigateTo(m Model, text string) Model {
	for i, t := range m.todos {
		if t.Text == text {
			m.cursor = i
			return m
		}
	}
	return m // no-op if not found
}

// markDone toggles a todo to done state (pending → in_progress → done).
func markDone(m Model, text string) Model {
	m = navigateTo(m, text)
	m = sendKey(m, "x") // pending → in_progress
	m = sendKey(m, "x") // in_progress → done
	return m
}

// ---- Basic model tests ----

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

// ---- Issue #4: Creation & Editing ----

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
	m = createTodo(m, "Old text #work")
	originalID := m.todos[0].ID

	m = sendKey(m, "e")
	if m.inputMode != inputModeEdit {
		t.Fatal("expected inputModeEdit after pressing e")
	}
	if m.editingID != originalID {
		t.Errorf("editingID mismatch")
	}
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
	m = createTodo(m, "First")
	m = createTodo(m, "Second")

	if m.cursor != 1 {
		t.Errorf("cursor should be at last created (1), got %d", m.cursor)
	}
	m = sendKey(m, "k")
	if m.cursor != 0 {
		t.Errorf("expected cursor 0 after k, got %d", m.cursor)
	}
	m = sendKey(m, "k") // clamp
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0, got %d", m.cursor)
	}
	m = sendKey(m, "j")
	if m.cursor != 1 {
		t.Errorf("expected cursor 1 after j, got %d", m.cursor)
	}
	m = sendKey(m, "j") // clamp
	if m.cursor != 1 {
		t.Errorf("cursor should stay at 1, got %d", m.cursor)
	}
}

// ---- Issue #5: Toggle, Delete & Undo ----

func TestToggleCycleStates(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Toggle me")

	// pending → in_progress
	m = sendKey(m, "x")
	if m.todos[0].GetState() != model.StateInProgress {
		t.Errorf("expected StateInProgress after first toggle, got %v", m.todos[0].GetState())
	}

	// in_progress → done
	m = sendKey(m, "x")
	if m.todos[0].GetState() != model.StateDone {
		t.Errorf("expected StateDone after second toggle, got %v", m.todos[0].GetState())
	}
	if m.todos[0].CompletedAt == nil {
		t.Error("CompletedAt should be set when done")
	}

	// done → pending
	m = sendKey(m, "x")
	if m.todos[0].GetState() != model.StatePending {
		t.Errorf("expected StatePending after third toggle, got %v", m.todos[0].GetState())
	}
	if m.todos[0].CompletedAt != nil {
		t.Error("CompletedAt should be cleared when back to pending")
	}
}

func TestDeleteDoneTodoImmediately(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Done item")
	// Toggle to done (pending → in_progress → done).
	m = sendKey(m, "x")
	m = sendKey(m, "x")

	m = sendKey(m, "d")
	// No confirmation dialog should appear for done todos.
	if m.showConfirm {
		t.Error("should not show confirmation for done todo")
	}
	if len(m.todos) != 0 {
		t.Errorf("expected 0 todos after deleting done todo, got %d", len(m.todos))
	}
}

func TestDeleteIncompleteTodoShowsConfirm(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Pending item")

	m = sendKey(m, "d")
	if !m.showConfirm {
		t.Fatal("expected confirmation dialog for incomplete todo")
	}
	if m.confirmTodoIdx != 0 {
		t.Errorf("confirmTodoIdx should be 0, got %d", m.confirmTodoIdx)
	}
	if len(m.todos) != 1 {
		t.Error("todo should not be deleted yet (dialog shown)")
	}
}

func TestDeleteConfirmYes(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Pending item")
	m = sendKey(m, "d") // show confirm
	m = sendKey(m, "y") // confirm

	if m.showConfirm {
		t.Error("dialog should be dismissed")
	}
	if len(m.todos) != 0 {
		t.Errorf("expected 0 todos after confirming delete, got %d", len(m.todos))
	}
}

func TestDeleteConfirmNo(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Pending item")
	m = sendKey(m, "d") // show confirm
	m = sendKey(m, "n") // cancel

	if m.showConfirm {
		t.Error("dialog should be dismissed")
	}
	if len(m.todos) != 1 {
		t.Errorf("expected 1 todo after cancelling delete, got %d", len(m.todos))
	}
}

func TestDeleteConfirmEsc(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Pending item")
	m = sendKey(m, "d")               // show confirm
	m = sendSpecialKey(m, tea.KeyEsc) // cancel with Esc

	if m.showConfirm {
		t.Error("dialog should be dismissed after Esc")
	}
	if len(m.todos) != 1 {
		t.Error("todo should not be deleted after Esc")
	}
}

func TestDeleteAllCompleted(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Todo 1")
	m = createTodo(m, "Todo 2 done")
	m = createTodo(m, "Todo 3 done")

	// Mark "Todo 2 done" and "Todo 3 done" as done using named navigation
	// so that sort-induced cursor movement does not break the test.
	m = markDone(m, "Todo 2 done")
	m = markDone(m, "Todo 3 done")

	m = sendKey(m, "D")

	if len(m.todos) != 1 {
		t.Errorf("expected 1 todo after D, got %d", len(m.todos))
	}
	if m.todos[0].Text != "Todo 1" {
		t.Errorf("expected 'Todo 1' to remain, got %q", m.todos[0].Text)
	}
	if len(m.undoStack) != 2 {
		t.Errorf("expected 2 undo entries, got %d", len(m.undoStack))
	}
}

func TestUndoDeleteRestoresTodo(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Restore me #work")
	// Toggle to done and delete.
	m = sendKey(m, "x")
	m = sendKey(m, "x")
	m = sendKey(m, "d")

	if len(m.todos) != 0 {
		t.Fatal("todo should be deleted before undo")
	}

	m = sendKey(m, "u")
	if len(m.todos) != 1 {
		t.Fatalf("expected 1 todo after undo, got %d", len(m.todos))
	}
	if m.todos[0].Text != "Restore me #work" {
		t.Errorf("unexpected text after undo: %q", m.todos[0].Text)
	}
	if m.statusMsg != "Todo restored" {
		t.Errorf("expected 'Todo restored' status, got %q", m.statusMsg)
	}
}

func TestUndoMultiple(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "First")
	m = createTodo(m, "Second")

	// Delete both (toggle to done first using named navigation).
	m = markDone(m, "First")
	m = navigateTo(m, "First")
	m = sendKey(m, "d")

	m = markDone(m, "Second")
	m = navigateTo(m, "Second")
	m = sendKey(m, "d")

	if len(m.todos) != 0 {
		t.Fatalf("expected 0 todos after two deletes, got %d", len(m.todos))
	}

	// Undo twice.
	m = sendKey(m, "u")
	if len(m.todos) != 1 {
		t.Fatalf("expected 1 todo after first undo, got %d", len(m.todos))
	}

	m = sendKey(m, "u")
	if len(m.todos) != 2 {
		t.Fatalf("expected 2 todos after second undo, got %d", len(m.todos))
	}
}

func TestUndoStackEmpty(t *testing.T) {
	m := tempModel(t)
	// Pressing u with nothing to undo should be a no-op.
	m = sendKey(m, "u")
	if len(m.todos) != 0 {
		t.Error("pressing u with empty undo stack should be a no-op")
	}
}

func TestViewRenders(t *testing.T) {
	m := tempModel(t)
	v := m.View()
	if v == "" {
		t.Fatal("View should not return empty string")
	}
}

func TestHelpWindowToggle(t *testing.T) {
	m := tempModel(t)

	if m.showHelp {
		t.Fatal("help should be hidden initially")
	}

	m = sendKey(m, "?")
	if !m.showHelp {
		t.Fatal("? should open help window")
	}

	// While help is open, other keys (like q) close it — not quit.
	m = sendKey(m, "q")
	if m.showHelp {
		t.Fatal("q should close help window")
	}

	// Re-open and close with ?
	m = sendKey(m, "?")
	m = sendKey(m, "?")
	if m.showHelp {
		t.Fatal("second ? should close help window")
	}
}

func TestHelpWindowBlocksInput(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "existing")
	m = sendKey(m, "?") // open help

	// Pressing i while help is open should not activate input.
	m = sendKey(m, "i")
	if m.inputMode != inputModeNone {
		t.Error("input should not activate while help window is open")
	}
}

func TestSortOnToggle(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "A")
	m = createTodo(m, "B")

	// Toggle A to done — it should move after B.
	m = markDone(m, "A")

	if m.todos[0].Text != "B" {
		t.Errorf("expected B (pending) first after A done, got %q", m.todos[0].Text)
	}
	if m.todos[1].Text != "A" {
		t.Errorf("expected A (done) second, got %q", m.todos[1].Text)
	}
}

func TestSortOnCreate(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "First")
	// Mark First as done so Second will appear before it.
	m = markDone(m, "First")
	m = createTodo(m, "Second")

	// "Second" is pending so should appear before "First" (done).
	if m.todos[0].Text != "Second" {
		t.Errorf("expected Second (pending) first, got %q", m.todos[0].Text)
	}
}
