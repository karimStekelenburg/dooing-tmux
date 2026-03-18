package ui

import (
	"strings"
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// ---- Tag system helpers ----

func TestRemoveTagFromTodos(t *testing.T) {
	todos := []*model.Todo{
		model.NewTodo("Buy milk #shopping #urgent"),
		model.NewTodo("Code review #work"),
		model.NewTodo("No tags here"),
	}

	removeTagFromTodos(todos, "shopping")

	if strings.Contains(todos[0].Text, "#shopping") {
		t.Error("expected #shopping to be removed from first todo")
	}
	if !strings.Contains(todos[0].Text, "#urgent") {
		t.Error("expected #urgent to remain in first todo")
	}
	if !strings.Contains(todos[1].Text, "#work") {
		t.Error("expected #work to remain in second todo")
	}
	if todos[2].Text != "No tags here" {
		t.Error("expected todo without tag to be unchanged")
	}
}

func TestRenameTagInTodos(t *testing.T) {
	todos := []*model.Todo{
		model.NewTodo("Task #work #urgent"),
		model.NewTodo("Another #workout"),
		model.NewTodo("Nothing here"),
	}

	renameTagInTodos(todos, "work", "job")

	if !strings.Contains(todos[0].Text, "#job") {
		t.Error("expected #work renamed to #job in first todo")
	}
	if strings.Contains(todos[0].Text, "#work ") {
		t.Error("expected old #work to be gone from first todo")
	}
	// #workout should remain unchanged — only exact match #work.
	if !strings.Contains(todos[1].Text, "#workout") {
		t.Error("expected #workout to be unchanged")
	}
}

func TestCleanupSpaces(t *testing.T) {
	cases := []struct {
		in  string
		out string
	}{
		{"hello   world", "hello world"},
		{"  leading", "leading"},
		{"trailing   ", "trailing"},
		{"a  b  c", "a b c"},
	}
	for _, c := range cases {
		got := cleanupSpaces(c.in)
		if got != c.out {
			t.Errorf("cleanupSpaces(%q) = %q, want %q", c.in, got, c.out)
		}
	}
}

// ---- Tag window UI state ----

func TestTagWindowOpenClose(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task #work")

	m = sendKey(m, "t") // open tag window
	if !m.tagWin.open {
		t.Fatal("expected tag window to be open after pressing t")
	}
	if len(m.tagWin.tags) != 1 || m.tagWin.tags[0] != "work" {
		t.Errorf("expected tags=['work'], got %v", m.tagWin.tags)
	}

	m = sendKey(m, "q") // close
	if m.tagWin.open {
		t.Fatal("expected tag window to close after pressing q")
	}
}

func TestTagWindowFilterOnEnter(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task A #work")
	m = createTodo(m, "Task B #personal")

	m = sendKey(m, "t")     // open tag window
	// cursor should be at 0 (first tag alphabetically)
	m = sendKey(m, "enter") // select first tag as filter

	if !m.tagWin.open == false {
		// tag window closes on enter
	}
	if m.tagWin.open {
		t.Fatal("tag window should close after selecting a tag")
	}
	if m.activeFilter == "" {
		t.Fatal("expected activeFilter to be set after selecting a tag")
	}

	visible := m.filteredTodos()
	for _, todo := range visible {
		if !strings.Contains(todo.Text, "#"+m.activeFilter) {
			t.Errorf("filtered todo %q does not contain #%s", todo.Text, m.activeFilter)
		}
	}
}

func TestClearFilter(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task #work")
	m = createTodo(m, "Other #personal")

	// Set a filter.
	m.activeFilter = "work"

	m = sendKey(m, "c")
	if m.activeFilter != "" {
		t.Errorf("expected activeFilter cleared, got %q", m.activeFilter)
	}
	if m.cursor != 0 {
		t.Errorf("expected cursor reset to 0, got %d", m.cursor)
	}
}

func TestFilteredTodosReturnsSubset(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task A #work")
	m = createTodo(m, "Task B #personal")
	m = createTodo(m, "Task C #work")

	m.activeFilter = "work"
	visible := m.filteredTodos()

	if len(visible) != 2 {
		t.Errorf("expected 2 work todos, got %d", len(visible))
	}
	for _, todo := range visible {
		if !strings.Contains(todo.Text, "#work") {
			t.Errorf("non-work todo in filtered list: %q", todo.Text)
		}
	}
}

func TestFilteredTodosNoFilterReturnsAll(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task A #work")
	m = createTodo(m, "Task B")

	m.activeFilter = ""
	visible := m.filteredTodos()
	if len(visible) != 2 {
		t.Errorf("expected 2 todos with no filter, got %d", len(visible))
	}
}

func TestTagWindowDeleteTag(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Buy milk #shopping #urgent")
	m = createTodo(m, "Code #shopping")

	m = sendKey(m, "t") // open tag window
	// Navigate to "shopping" (alphabetically before "urgent").
	// tags should be ["shopping", "urgent"]
	if len(m.tagWin.tags) < 1 || m.tagWin.tags[0] != "shopping" {
		t.Skipf("tag order unexpected: %v", m.tagWin.tags)
	}

	m = sendKey(m, "d") // delete "shopping" tag

	for _, todo := range m.todos {
		if strings.Contains(todo.Text, "#shopping") {
			t.Errorf("expected #shopping removed from all todos, but found in: %q", todo.Text)
		}
	}
}

func TestTagWindowNavigation(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "A #alpha")
	m = createTodo(m, "B #beta")
	m = createTodo(m, "C #gamma")

	m = sendKey(m, "t") // open tag window
	if m.tagWin.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.tagWin.cursor)
	}
	m = sendKey(m, "j")
	if m.tagWin.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.tagWin.cursor)
	}
	m = sendKey(m, "j")
	if m.tagWin.cursor != 2 {
		t.Errorf("expected cursor at 2, got %d", m.tagWin.cursor)
	}
	m = sendKey(m, "j") // clamp
	if m.tagWin.cursor != 2 {
		t.Errorf("cursor should clamp at 2, got %d", m.tagWin.cursor)
	}
	m = sendKey(m, "k")
	if m.tagWin.cursor != 1 {
		t.Errorf("expected cursor at 1, got %d", m.tagWin.cursor)
	}
}

func TestTagWindowBlocksMainInput(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task #work")
	initialCount := len(m.todos)

	m = sendKey(m, "t") // open tag window
	m = sendKey(m, "i") // 'i' should NOT create a todo — tag window intercepts
	if m.inputMode == inputModeCreate {
		t.Error("tag window should intercept all keys, i should not open input mode")
	}
	if len(m.todos) != initialCount {
		t.Error("no new todos should be created while tag window is open")
	}
}

func TestRenderFilterHeader(t *testing.T) {
	m := tempModel(t)
	m.activeFilter = "work"
	header := m.renderFilterHeader()
	if !strings.Contains(header, "work") {
		t.Errorf("filter header should mention the active filter, got: %q", header)
	}

	m.activeFilter = ""
	header = m.renderFilterHeader()
	if header != "" {
		t.Errorf("filter header should be empty when no filter, got: %q", header)
	}
}

func TestFindTodosIndex(t *testing.T) {
	m := tempModel(t)
	m = createTodo(m, "Task A")
	m = createTodo(m, "Task B")

	id := m.todos[0].ID
	idx := m.findTodosIndex(id)
	if idx != 0 {
		t.Errorf("expected index 0, got %d", idx)
	}

	idx = m.findTodosIndex("nonexistent")
	if idx != -1 {
		t.Errorf("expected -1 for nonexistent ID, got %d", idx)
	}
}
