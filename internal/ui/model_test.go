package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewModel(t *testing.T) {
	m := NewModel(false)
	if m.projectMode {
		t.Error("expected projectMode=false")
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

func TestViewRendersEmptyState(t *testing.T) {
	m := NewModel(false)
	m.todos = nil
	view := m.View()
	if len(view) == 0 {
		t.Error("expected non-empty view")
	}
}

func TestCursorNavigation(t *testing.T) {
	m := NewModel(false)

	// Inject todos directly
	m.todos = makeTodos(3)

	// j moves down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.cursor != 1 {
		t.Errorf("expected cursor=1 after j, got %d", um.cursor)
	}

	// k moves up
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um = updated.(Model)
	if um.cursor != 0 {
		t.Errorf("expected cursor=0 after k, got %d", um.cursor)
	}

	// G goes to last
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	um = updated.(Model)
	if um.cursor != 2 {
		t.Errorf("expected cursor=2 after G, got %d", um.cursor)
	}

	// g goes to first
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	um = updated.(Model)
	if um.cursor != 0 {
		t.Errorf("expected cursor=0 after g, got %d", um.cursor)
	}
}

func TestCursorClamping(t *testing.T) {
	m := NewModel(false)
	m.todos = makeTodos(2)

	// Can't go above 0
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um := updated.(Model)
	if um.cursor != 0 {
		t.Errorf("cursor should clamp at 0")
	}

	// Move to last, can't go past
	um.cursor = 1
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um = updated.(Model)
	if um.cursor != 1 {
		t.Errorf("cursor should clamp at last index")
	}
}

func TestRelativeTime(t *testing.T) {
	now := int64(0) // use Unix epoch for testing — but relativeTime uses time.Since
	_ = now
	// Just verify it returns a non-empty string
	r := relativeTime(0)
	if r == "" {
		t.Error("expected non-empty relative time")
	}
}
