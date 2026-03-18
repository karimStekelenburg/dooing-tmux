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
	"github.com/karimStekelenburg/dooing-tmux/internal/priority"
)

// Sort sorts todos in-place using the multi-key comparator.
// It is stable: equal items preserve their relative insertion order.
//
// cfg parameters:
//   - doneSortByCompleted: when true, among done todos sort by most-recently completed first.
func Sort(todos []*model.Todo, doneSortByCompleted bool) {
	SortWithConfig(todos, doneSortByCompleted, config.Defaults())
}

// SortWithConfig sorts todos using priority scoring from the provided config.
func SortWithConfig(todos []*model.Todo, doneSortByCompleted bool, cfg config.Config) {
	sort.SliceStable(todos, func(i, j int) bool {
		return less(todos[i], todos[j], doneSortByCompleted, cfg)
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
	aScore := priority.GetScore(a, cfg)
	bScore := priority.GetScore(b, cfg)
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

// completedTime returns CompletedAt as an int64 (0 if nil).
func completedTime(t *model.Todo) int64 {
	if t.CompletedAt == nil {
		return 0
	}
	return *t.CompletedAt
}
