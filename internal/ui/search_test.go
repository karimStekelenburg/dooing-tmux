package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func fakeKeySearch(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func TestRunSearch(t *testing.T) {
	m := NewModel(false)
	t1 := model.NewTodo("Buy groceries")
	t2 := model.NewTodo("Write tests")
	t3 := model.NewTodo("Buy milk")
	m.todos = []*model.Todo{t1, t2, t3}

	m.search.input.SetValue("buy")
	m.runSearch()

	if len(m.search.results) != 2 {
		t.Fatalf("expected 2 results for 'buy', got %d", len(m.search.results))
	}
}

func TestRunSearchCaseInsensitive(t *testing.T) {
	m := NewModel(false)
	t1 := model.NewTodo("HELLO WORLD")
	m.todos = []*model.Todo{t1}

	m.search.input.SetValue("hello")
	m.runSearch()

	if len(m.search.results) != 1 {
		t.Fatalf("expected 1 result for case-insensitive 'hello', got %d", len(m.search.results))
	}
}

func TestRunSearchEmpty(t *testing.T) {
	m := NewModel(false)
	t1 := model.NewTodo("something")
	m.todos = []*model.Todo{t1}

	m.search.input.SetValue("")
	m.runSearch()

	if len(m.search.results) != 0 {
		t.Fatalf("expected 0 results for empty query, got %d", len(m.search.results))
	}
}

func TestRunSearchNoMatches(t *testing.T) {
	m := NewModel(false)
	t1 := model.NewTodo("hello world")
	m.todos = []*model.Todo{t1}

	m.search.input.SetValue("xyz")
	m.runSearch()

	if len(m.search.results) != 0 {
		t.Fatalf("expected 0 results for 'xyz', got %d", len(m.search.results))
	}
}

func TestSearchEnterJumpsToTodo(t *testing.T) {
	m := NewModel(false)
	t1 := model.NewTodo("first")
	t2 := model.NewTodo("second match")
	t3 := model.NewTodo("third")
	m.todos = []*model.Todo{t1, t2, t3}
	m.cursor = 0
	m.search.open = true
	m.search.results = []*model.Todo{t2}
	m.search.cursor = 0

	result, _ := m.updateSearch(fakeKeySearch("enter"))
	updated := result.(Model)

	if updated.search.open {
		t.Error("expected search to be closed after enter")
	}
	if updated.cursor != 1 {
		t.Errorf("expected cursor at index 1 (second todo), got %d", updated.cursor)
	}
}
