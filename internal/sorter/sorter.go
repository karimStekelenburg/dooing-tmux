// Package sorter implements the multi-key stable sort for todos.
//
// Sort order (all keys applied left-to-right):
//  1. Completion status: undone (pending/in-progress) before done.
//  2. Among done items: most-recently completed first (when doneSortByCompleted is true).
//  3. Priority score: higher score first (descending). Stub returns 0; full scoring in #8.
//  4. Due date: earlier first; todos WITH a due date before those WITHOUT.
//  5. Creation time: earlier first (ascending / stable insertion order).
package sorter

import (
	"sort"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// Sort sorts todos in-place using the multi-key comparator.
// It is stable: equal items preserve their relative insertion order.
//
// cfg parameters:
//   - doneSortByCompleted: when true, among done todos sort by most-recently completed first.
// Sort sorts todos in-place using the multi-key comparator.
// cfg is used for priority scoring; pass nil to use zero scores.
func Sort(todos []*model.Todo, doneSortByCompleted bool, cfg ...config.Config) {
	var c config.Config
	if len(cfg) > 0 {
		c = cfg[0]
	}
	sort.SliceStable(todos, func(i, j int) bool {
		return less(todos[i], todos[j], doneSortByCompleted, c)
	})
}

// less is the multi-key comparator used by Sort.
func less(a, b *model.Todo, doneSortByCompleted bool, cfg config.Config) bool {
	aDone := a.Done
	bDone := b.Done

	// 1. Undone before done.
	if aDone != bDone {
		return !aDone // undone (false) < done (true)
	}

	// 2. Among done: most-recently completed first.
	if aDone && doneSortByCompleted {
		aTime := completedTime(a)
		bTime := completedTime(b)
		if aTime != bTime {
			return aTime > bTime // larger unix timestamp = more recent
		}
	}

	// 3. Priority score: higher first.
	aScore := GetPriorityScore(a, cfg)
	bScore := GetPriorityScore(b, cfg)
	if aScore != bScore {
		return aScore > bScore
	}

	// 4. Due date: earlier first; WITH due date before WITHOUT.
	aHasDue := a.DueAt != nil
	bHasDue := b.DueAt != nil
	if aHasDue != bHasDue {
		return aHasDue // with-due before without-due
	}
	if aHasDue && bHasDue {
		if *a.DueAt != *b.DueAt {
			return *a.DueAt < *b.DueAt
		}
	}

	// 5. Creation time: earlier first (preserves insertion order for new todos).
	return a.CreatedAt < b.CreatedAt
}

// GetPriorityScore computes a priority score for a todo.
// Done todos always score 0.
// base = sum of weights for each priority the todo has.
// multiplier = 1.0 if no estimated hours, else 1.0 / (hours * hourScoreValue).
// score = base * multiplier.
func GetPriorityScore(t *model.Todo, cfg config.Config) float64 {
	if t.Done {
		return 0
	}
	if len(t.Priorities) == 0 {
		return 0
	}

	// Build weight lookup.
	weightMap := make(map[string]int, len(cfg.Priorities))
	for _, p := range cfg.Priorities {
		weightMap[p.Name] = p.Weight
	}

	var base float64
	for _, name := range t.Priorities {
		base += float64(weightMap[name])
	}

	multiplier := 1.0
	if t.EstimatedHours > 0 {
		hsv := cfg.HourScoreValue
		if hsv <= 0 {
			hsv = 0.125
		}
		multiplier = 1.0 / (t.EstimatedHours * hsv)
	}

	return base * multiplier
}

// completedTime returns CompletedAt as an int64 (0 if nil).
func completedTime(t *model.Todo) int64 {
	if t.CompletedAt == nil {
		return 0
	}
	return *t.CompletedAt
}
