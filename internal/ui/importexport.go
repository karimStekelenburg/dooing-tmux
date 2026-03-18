package ui

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// importTodos reads todos from the given file path, appends them to existing,
// deduplicates, and returns the merged slice.
func importTodos(existing []*model.Todo, filePath string) ([]*model.Todo, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", filePath)
		}
		return nil, fmt.Errorf("read error: %w", err)
	}

	var incoming []*model.Todo
	if err := json.Unmarshal(data, &incoming); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	merged := append(existing, incoming...)
	return deduplicateTodos(merged), nil
}

// exportTodos writes todos as JSON to the given file path.
func exportTodos(todos []*model.Todo, filePath string) error {
	data, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}
	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("write error: %w", err)
	}
	return nil
}

// deduplicateTodos removes duplicate todos using a SHA-256 hash of canonical
// JSON (all fields except ID and CreatedAt). The first occurrence is kept.
func deduplicateTodos(todos []*model.Todo) []*model.Todo {
	seen := make(map[[32]byte]bool, len(todos))
	result := make([]*model.Todo, 0, len(todos))
	for _, t := range todos {
		h := todoHash(t)
		if !seen[h] {
			seen[h] = true
			result = append(result, t)
		}
	}
	return result
}

// todoHash computes a SHA-256 hash of a todo's canonical representation.
// ID and CreatedAt are excluded so that semantically identical todos
// (same text, state, tags, etc.) are considered duplicates even if they were
// created at different times or imported from different sources.
type canonicalTodo struct {
	Text           string   `json:"text"`
	Done           bool     `json:"done"`
	InProgress     bool     `json:"in_progress"`
	Category       string   `json:"category"`
	Priorities     []string `json:"priorities,omitempty"`
	EstimatedHours float64  `json:"estimated_hours,omitempty"`
	DueAt          *int64   `json:"due_at,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	ParentID       string   `json:"parent_id,omitempty"`
	Depth          int      `json:"depth,omitempty"`
}

func todoHash(t *model.Todo) [32]byte {
	c := canonicalTodo{
		Text:           t.Text,
		Done:           t.Done,
		InProgress:     t.InProgress,
		Category:       t.Category,
		Priorities:     t.Priorities,
		EstimatedHours: t.EstimatedHours,
		DueAt:          t.DueAt,
		Notes:          t.Notes,
		ParentID:       t.ParentID,
		Depth:          t.Depth,
	}
	data, _ := json.Marshal(c)
	return sha256.Sum256(data)
}
