package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewTodo(t *testing.T) {
	todo := NewTodo("buy groceries #work")
	if todo.ID == "" {
		t.Error("expected non-empty ID")
	}
	if todo.Text != "buy groceries #work" {
		t.Errorf("unexpected text: %s", todo.Text)
	}
	if todo.Category != "work" {
		t.Errorf("expected category 'work', got '%s'", todo.Category)
	}
	if todo.CreatedAt == 0 {
		t.Error("expected non-zero CreatedAt")
	}
}

func TestToggleStateMachine(t *testing.T) {
	todo := NewTodo("test")

	// pending → in_progress
	if todo.GetState() != StatePending {
		t.Error("initial state should be pending")
	}
	todo.Toggle()
	if todo.GetState() != StateInProgress {
		t.Error("after 1st toggle should be in_progress")
	}
	if todo.CompletedAt != nil {
		t.Error("CompletedAt should be nil in in_progress")
	}

	// in_progress → done
	todo.Toggle()
	if todo.GetState() != StateDone {
		t.Error("after 2nd toggle should be done")
	}
	if todo.CompletedAt == nil {
		t.Error("CompletedAt should be set when done")
	}

	// done → pending
	todo.Toggle()
	if todo.GetState() != StatePending {
		t.Error("after 3rd toggle should be pending again")
	}
	if todo.CompletedAt != nil {
		t.Error("CompletedAt should be cleared on return to pending")
	}
}

func TestJSONRoundTrip(t *testing.T) {
	original := NewTodo("buy milk #shopping")
	original.Toggle() // pending → in_progress
	original.Toggle() // in_progress → done

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var restored Todo
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID mismatch: %s vs %s", restored.ID, original.ID)
	}
	if restored.Text != original.Text {
		t.Errorf("Text mismatch")
	}
	if restored.Done != original.Done {
		t.Errorf("Done mismatch")
	}
	if restored.CompletedAt == nil || *restored.CompletedAt != *original.CompletedAt {
		t.Errorf("CompletedAt mismatch")
	}
}

func TestExtractAllTags(t *testing.T) {
	todo := &Todo{Text: "do something #work #urgent #work"}
	tags := todo.ExtractAllTags()
	if len(tags) != 2 {
		t.Errorf("expected 2 unique tags, got %d: %v", len(tags), tags)
	}
}

func TestIsOverdue(t *testing.T) {
	past := time.Now().Add(-48 * time.Hour).Unix()
	future := time.Now().Add(48 * time.Hour).Unix()

	overdue := &Todo{DueAt: &past}
	if !overdue.IsOverdue() {
		t.Error("expected overdue=true for past due date")
	}

	notDue := &Todo{DueAt: &future}
	if notDue.IsOverdue() {
		t.Error("expected overdue=false for future due date")
	}

	done := &Todo{Done: true, DueAt: &past}
	if done.IsOverdue() {
		t.Error("done todos should not be overdue")
	}
}

func TestGetAllTags(t *testing.T) {
	todos := []*Todo{
		{Text: "task #alpha"},
		{Text: "task #beta #alpha"},
		{Text: "task #gamma"},
	}
	tags := GetAllTags(todos)
	if len(tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(tags), tags)
	}
	// Should be sorted
	if tags[0] != "alpha" || tags[1] != "beta" || tags[2] != "gamma" {
		t.Errorf("unexpected order: %v", tags)
	}
}
