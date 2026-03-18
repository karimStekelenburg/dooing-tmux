package model

import (
	"fmt"
	"math/rand"
	"time"
)

// State represents the tri-state status of a todo.
type State int

const (
	StatePending    State = iota // ○
	StateInProgress              // ◐
	StateDone                    // ✓
)

// Todo is the core data structure for a single todo item.
type Todo struct {
	ID             string   `json:"id"`
	Text           string   `json:"text"`
	Done           bool     `json:"done"`
	InProgress     bool     `json:"in_progress"`
	Category       string   `json:"category"`
	CreatedAt      int64    `json:"created_at"`
	CompletedAt    *int64   `json:"completed_at,omitempty"`
	Priorities     []string `json:"priorities,omitempty"`
	EstimatedHours float64  `json:"estimated_hours,omitempty"`
	DueAt          *int64   `json:"due_at,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	ParentID       string   `json:"parent_id,omitempty"`
	Depth          int      `json:"depth,omitempty"`
}

// NewTodo creates a new Todo with a generated ID and the current timestamp.
func NewTodo(text string) *Todo {
	now := time.Now().Unix()
	id := fmt.Sprintf("%d_%d", now, rand.Intn(9000)+1000) //nolint:gosec
	t := &Todo{
		ID:        id,
		Text:      text,
		CreatedAt: now,
	}
	t.Category = t.ExtractCategory()
	return t
}

// State returns the tri-state status of the todo.
func (t *Todo) GetState() State {
	if t.Done {
		return StateDone
	}
	if t.InProgress {
		return StateInProgress
	}
	return StatePending
}

// Toggle cycles the todo through: pending → in_progress → done → pending.
// It also manages CompletedAt.
func (t *Todo) Toggle() {
	switch t.GetState() {
	case StatePending:
		t.InProgress = true
		t.Done = false
	case StateInProgress:
		t.InProgress = false
		t.Done = true
		now := time.Now().Unix()
		t.CompletedAt = &now
	case StateDone:
		t.Done = false
		t.InProgress = false
		t.CompletedAt = nil
	}
}

// ExtractCategory returns the first #tag found in the todo text, or "".
func (t *Todo) ExtractCategory() string {
	tags := extractTags(t.Text)
	if len(tags) == 0 {
		return ""
	}
	return tags[0]
}

// ExtractAllTags returns all #tags found in the todo text.
func (t *Todo) ExtractAllTags() []string {
	return extractTags(t.Text)
}

// IsOverdue returns true if the todo has a due date in the past and is not done.
func (t *Todo) IsOverdue() bool {
	if t.Done || t.DueAt == nil {
		return false
	}
	return *t.DueAt < startOfToday()
}

func startOfToday() int64 {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
}
