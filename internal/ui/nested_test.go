package ui

import (
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func TestHasChildren(t *testing.T) {
	parent := model.NewTodo("parent")
	child := model.NewTodo("child")
	child.ParentID = parent.ID
	child.Depth = 1
	other := model.NewTodo("other")

	todos := []*model.Todo{parent, child, other}

	if !hasChildren(todos, parent.ID) {
		t.Error("expected hasChildren to return true for parent")
	}
	if hasChildren(todos, other.ID) {
		t.Error("expected hasChildren to return false for other")
	}
}

func TestCountDescendants(t *testing.T) {
	root := model.NewTodo("root")
	child1 := model.NewTodo("child1")
	child1.ParentID = root.ID
	child1.Depth = 1
	child2 := model.NewTodo("child2")
	child2.ParentID = root.ID
	child2.Depth = 1
	grandchild := model.NewTodo("grandchild")
	grandchild.ParentID = child1.ID
	grandchild.Depth = 2

	todos := []*model.Todo{root, child1, child2, grandchild}
	n := countDescendants(todos, root.ID)
	if n != 3 {
		t.Errorf("expected 3 descendants, got %d", n)
	}
	n2 := countDescendants(todos, child1.ID)
	if n2 != 1 {
		t.Errorf("expected 1 descendant of child1, got %d", n2)
	}
}

func TestPromoteOrphans(t *testing.T) {
	parent := model.NewTodo("parent")
	child := model.NewTodo("child")
	child.ParentID = "nonexistent"
	child.Depth = 1

	todos := []*model.Todo{parent, child}
	promoteOrphans(todos)

	if child.ParentID != "" {
		t.Errorf("expected child.ParentID to be empty after promotion, got %q", child.ParentID)
	}
	if child.Depth != 0 {
		t.Errorf("expected child.Depth to be 0 after promotion, got %d", child.Depth)
	}
}

func TestIsDescendant(t *testing.T) {
	root := model.NewTodo("root")
	child := model.NewTodo("child")
	child.ParentID = root.ID
	grandchild := model.NewTodo("grandchild")
	grandchild.ParentID = child.ID

	todos := []*model.Todo{root, child, grandchild}

	if !isDescendant(todos, grandchild.ID, root.ID) {
		t.Error("grandchild should be a descendant of root")
	}
	if !isDescendant(todos, child.ID, root.ID) {
		t.Error("child should be a descendant of root")
	}
	if isDescendant(todos, root.ID, grandchild.ID) {
		t.Error("root should NOT be a descendant of grandchild")
	}
}

func TestVisibleTodosWithFold(t *testing.T) {
	m := NewModel(false)
	parent := model.NewTodo("parent")
	child := model.NewTodo("child")
	child.ParentID = parent.ID
	child.Depth = 1

	m.todos = []*model.Todo{parent, child}

	// Before fold: both visible.
	vis := m.visibleTodos()
	if len(vis) != 2 {
		t.Fatalf("expected 2 visible todos before fold, got %d", len(vis))
	}

	// After fold: only parent visible.
	m.nested.toggleFold(parent.ID)
	vis = m.visibleTodos()
	if len(vis) != 1 {
		t.Fatalf("expected 1 visible todo after fold, got %d", len(vis))
	}
	if vis[0].ID != parent.ID {
		t.Error("expected parent to be the only visible todo")
	}

	// After unfold: both visible again.
	m.nested.toggleFold(parent.ID)
	vis = m.visibleTodos()
	if len(vis) != 2 {
		t.Fatalf("expected 2 visible todos after unfold, got %d", len(vis))
	}
}
