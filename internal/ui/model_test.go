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
