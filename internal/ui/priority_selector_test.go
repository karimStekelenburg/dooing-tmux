package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func TestPrioritySelectorOpenClose(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task with priorities")

	m = sendKey(m, "p")
	if !m.priSel.open {
		t.Fatal("expected priority selector to open after pressing p")
	}
	if len(m.priSel.items) == 0 {
		t.Fatal("expected priority items to be populated from config")
	}

	m = sendKey(m, "q")
	if m.priSel.open {
		t.Fatal("expected priority selector to close after q")
	}
}

func TestPrioritySelectorEscCloses(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task")
	m = sendKey(m, "p")
	m = sendSpecialKey(m, tea.KeyEsc)
	if m.priSel.open {
		t.Fatal("expected priority selector to close after Esc")
	}
}

func TestPrioritySelectorNavigation(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task")
	m = sendKey(m, "p")

	if m.priSel.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.priSel.cursor)
	}
	m = sendKey(m, "j")
	if m.priSel.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.priSel.cursor)
	}
	m = sendKey(m, "k")
	if m.priSel.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.priSel.cursor)
	}
	// Clamp at 0.
	m = sendKey(m, "k")
	if m.priSel.cursor != 0 {
		t.Errorf("cursor should clamp at 0, got %d", m.priSel.cursor)
	}
}

func TestPrioritySelectorSpaceToggle(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task")
	m = sendKey(m, "p")

	// Initially nothing is checked.
	for i, c := range m.priSel.checked {
		if c {
			t.Errorf("expected all unchecked initially, item %d is checked", i)
		}
	}

	// Toggle first item.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updated.(Model)
	if !m.priSel.checked[0] {
		t.Error("expected first item to be checked after space")
	}

	// Toggle again — uncheck.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updated.(Model)
	if m.priSel.checked[0] {
		t.Error("expected first item to be unchecked after second space")
	}
}

func TestPrioritySelectorConfirm(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task")
	todoID := m.todos[0].ID

	m = sendKey(m, "p")
	// Toggle first priority (index 0 = "important").
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = updated.(Model)
	m = sendSpecialKey(m, tea.KeyEnter)

	if m.priSel.open {
		t.Fatal("expected selector to close after enter")
	}

	// Find the todo and check priorities.
	var todo *model.Todo
	for _, t2 := range m.todos {
		if t2.ID == todoID {
			todo = t2
			break
		}
	}
	if todo == nil {
		t.Fatal("todo not found")
	}
	if len(todo.Priorities) == 0 {
		t.Fatal("expected priorities to be set after confirm")
	}
	if todo.Priorities[0] != "important" {
		t.Errorf("expected 'important', got %q", todo.Priorities[0])
	}
}

func TestPrioritySelectorBlocksMainInput(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task")
	m = sendKey(m, "p") // open selector

	// Pressing i while selector is open should not open input mode.
	m = sendKey(m, "i")
	if m.inputMode == inputModeCreate {
		t.Error("input mode should not activate while priority selector is open")
	}
}

func TestPriorityLabelFunction(t *testing.T) {
	t1 := model.NewTodo("Task")
	t1.Priorities = []string{"important", "urgent"}
	label := priorityLabel(t1)
	if !strings.Contains(label, "important") || !strings.Contains(label, "urgent") {
		t.Errorf("expected label to contain priority names, got %q", label)
	}

	t2 := model.NewTodo("No priorities")
	label = priorityLabel(t2)
	if label != "" {
		t.Errorf("expected empty label for todo with no priorities, got %q", label)
	}
}

func TestPrioritySelectorPrePopulated(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task with prio")
	// Set priority on the todo directly.
	m.todos[0].Priorities = []string{"urgent"}

	m = sendKey(m, "p")
	// Find "urgent" in items and check it is pre-checked.
	for i, name := range m.priSel.items {
		if name == "urgent" {
			if !m.priSel.checked[i] {
				t.Errorf("expected 'urgent' to be pre-checked, but it is not")
			}
		} else if name == "important" {
			if m.priSel.checked[i] {
				t.Errorf("expected 'important' to be unchecked, but it is checked")
			}
		}
	}
}
