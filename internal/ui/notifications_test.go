package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func todoWithDue(text string, daysOffset int) *model.Todo {
	t := model.NewTodo(text)
	now := time.Now()
	day := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())
	ts := day.AddDate(0, 0, daysOffset).Unix()
	t.DueAt = &ts
	return t
}

func TestNewNotifState_Empty(t *testing.T) {
	ns := newNotifState([]*model.Todo{})
	if ns.open {
		t.Error("expected notif to be closed when no todos")
	}
	if len(ns.items) != 0 {
		t.Errorf("expected 0 items, got %d", len(ns.items))
	}
}

func TestNewNotifState_Overdue(t *testing.T) {
	todos := []*model.Todo{
		todoWithDue("overdue task", -2),
	}
	ns := newNotifState(todos)
	if !ns.open {
		t.Error("expected notif to be open")
	}
	if len(ns.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(ns.items))
	}
	if !ns.items[0].overdue {
		t.Error("expected item to be overdue")
	}
}

func TestNewNotifState_DueToday(t *testing.T) {
	todos := []*model.Todo{
		todoWithDue("due today task", 0),
	}
	ns := newNotifState(todos)
	if !ns.open {
		t.Error("expected notif to be open")
	}
	if len(ns.items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(ns.items))
	}
	if ns.items[0].overdue {
		t.Error("expected item to be due today, not overdue")
	}
}

func TestNewNotifState_SkipsDone(t *testing.T) {
	overdue := todoWithDue("overdue but done", -2)
	overdue.Done = true
	dueToday := todoWithDue("due today but done", 0)
	dueToday.Done = true
	ns := newNotifState([]*model.Todo{overdue, dueToday})
	if ns.open {
		t.Error("expected notif to be closed for done todos")
	}
}

func TestNewNotifState_SkipsFuture(t *testing.T) {
	ns := newNotifState([]*model.Todo{
		todoWithDue("future task", 5),
	})
	if ns.open {
		t.Error("expected notif to be closed for future due dates")
	}
}

func TestUpdateNotifications_Navigation(t *testing.T) {
	m := Model{
		notif: notifState{
			open: true,
			items: []notifItem{
				{todo: model.NewTodo("a"), overdue: true},
				{todo: model.NewTodo("b"), overdue: false},
			},
			cursor: 0,
		},
	}

	// Move down.
	newM, _ := m.updateNotifications(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	m2 := newM.(Model)
	if m2.notif.cursor != 1 {
		t.Errorf("cursor after j: got %d, want 1", m2.notif.cursor)
	}

	// Move up.
	newM2, _ := m2.updateNotifications(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	m3 := newM2.(Model)
	if m3.notif.cursor != 0 {
		t.Errorf("cursor after k: got %d, want 0", m3.notif.cursor)
	}
}

func TestUpdateNotifications_Close(t *testing.T) {
	m := Model{
		notif: notifState{
			open:   true,
			items:  []notifItem{{todo: model.NewTodo("x"), overdue: true}},
			cursor: 0,
		},
	}

	for _, key := range []string{"q", "esc"} {
		var ktype tea.KeyType
		if key == "esc" {
			ktype = tea.KeyEsc
		} else {
			ktype = tea.KeyRunes
		}
		newM, _ := m.updateNotifications(tea.KeyMsg{Type: ktype, Runes: []rune(key)})
		m2 := newM.(Model)
		if m2.notif.open {
			t.Errorf("expected notif to close on %q", key)
		}
	}
}

func TestUpdateNotifications_Enter(t *testing.T) {
	todo := model.NewTodo("important task")
	m := Model{
		notif: notifState{
			open:   true,
			items:  []notifItem{{todo: todo, overdue: true}},
			cursor: 0,
		},
	}

	newM, cmd := m.updateNotifications(tea.KeyMsg{Type: tea.KeyEnter})
	m2 := newM.(Model)

	if m2.notif.open {
		t.Error("expected notif to close after Enter")
	}
	if cmd == nil {
		t.Fatal("expected a cmd to be returned")
	}

	// Execute the cmd to get the message.
	msg := cmd()
	jmp, ok := msg.(jumpToTodoMsg)
	if !ok {
		t.Fatalf("expected jumpToTodoMsg, got %T", msg)
	}
	if jmp.todoID != todo.ID {
		t.Errorf("jumpToTodoMsg.todoID = %q, want %q", jmp.todoID, todo.ID)
	}
}

func TestModel_JumpToTodoMsg(t *testing.T) {
	todos := []*model.Todo{
		model.NewTodo("first"),
		model.NewTodo("second"),
		model.NewTodo("third"),
	}
	m := Model{
		todos:  todos,
		cursor: 0,
	}

	newM, _ := m.Update(jumpToTodoMsg{todoID: todos[1].ID})
	m2 := newM.(Model)
	if m2.cursor != 1 {
		t.Errorf("cursor after jump: got %d, want 1", m2.cursor)
	}
}
