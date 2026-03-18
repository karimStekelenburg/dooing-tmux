package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func fakeSpecialKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

func TestOpenScratchpad(t *testing.T) {
	m := NewModel(false)
	todo := model.NewTodo("test todo")
	todo.Notes = "existing notes"
	m.todos = []*model.Todo{todo}
	m.cursor = 0
	m.width = 100
	m.height = 40

	cmd := m.openScratchpad()
	if cmd == nil {
		t.Error("expected a non-nil cmd from openScratchpad (textarea focus)")
	}
	if !m.pad.open {
		t.Error("expected pad.open to be true after openScratchpad")
	}
	if m.pad.todoID != todo.ID {
		t.Errorf("expected pad.todoID=%q, got %q", todo.ID, m.pad.todoID)
	}
	if m.pad.ta.Value() != "existing notes" {
		t.Errorf("expected textarea to be pre-filled with 'existing notes', got %q", m.pad.ta.Value())
	}
}

func TestScratchpadEscSavesNotes(t *testing.T) {
	m := NewModel(false)
	todo := model.NewTodo("test todo")
	m.todos = []*model.Todo{todo}
	m.cursor = 0
	m.width = 100
	m.height = 40
	m.pad.open = true
	m.pad.todoID = todo.ID
	m.pad.ta.SetValue("my new notes")

	result, _ := m.updateScratchpad(fakeSpecialKey(tea.KeyEsc))
	updated := result.(Model)

	if updated.pad.open {
		t.Error("expected pad to be closed after esc")
	}
	if updated.todos[0].Notes != "my new notes" {
		t.Errorf("expected notes to be saved, got %q", updated.todos[0].Notes)
	}
}

func TestNotesIconAppearsWhenNotesPresent(t *testing.T) {
	todo := model.NewTodo("test todo")
	todo.Notes = "some notes"

	line := renderTodo(todo, nil)
	if !containsRune(line, []rune(notesIcon)[0]) {
		t.Error("expected notes icon to appear in todo line when notes are set")
	}
}

func TestNotesIconAbsentWhenNoNotes(t *testing.T) {
	todo := model.NewTodo("test todo")
	todo.Notes = ""

	line := renderTodo(todo, nil)
	// Notes icon should NOT be present.
	for _, r := range line {
		for _, nr := range notesIcon {
			if r == nr {
				t.Error("expected notes icon to be absent when notes are empty")
				return
			}
		}
	}
}

func TestScratchpadEmptyNotesClearsIcon(t *testing.T) {
	m := NewModel(false)
	todo := model.NewTodo("test todo")
	todo.Notes = "old notes"
	m.todos = []*model.Todo{todo}
	m.cursor = 0
	m.width = 100
	m.height = 40
	m.pad.open = true
	m.pad.todoID = todo.ID
	m.pad.ta.SetValue("") // user cleared notes

	result, _ := m.updateScratchpad(fakeSpecialKey(tea.KeyEsc))
	updated := result.(Model)

	if updated.todos[0].Notes != "" {
		t.Errorf("expected notes to be cleared, got %q", updated.todos[0].Notes)
	}
}

// containsRune returns true if the string contains the given rune.
func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
