// Package priority implements priority scoring and color resolution for todos.
package priority

import (
	"sort"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// GetScore computes the priority score for a todo given the app config.
//
// Algorithm:
//   - Done todos always score 0.
//   - base = sum of weight for each priority name present in todo.Priorities.
//   - multiplier = 1.0 if EstimatedHours == 0, else 1.0 / (hours * cfg.HourScoreValue).
//   - score = base * multiplier.
func GetScore(todo *model.Todo, cfg config.Config) float64 {
	if todo.Done {
		return 0
	}

	// Build weight lookup.
	weightOf := make(map[string]int, len(cfg.Priorities))
	for _, p := range cfg.Priorities {
		weightOf[p.Name] = p.Weight
	}

	var base int
	for _, name := range todo.Priorities {
		base += weightOf[name]
	}

	if base == 0 {
		return 0
	}

	multiplier := 1.0
	if todo.EstimatedHours > 0 && cfg.HourScoreValue > 0 {
		multiplier = 1.0 / (todo.EstimatedHours * cfg.HourScoreValue)
	}

	return float64(base) * multiplier
}

// ResolveColor returns the color string for a todo based on its priorities
// and the configured priority groups.  Returns "" if no group matches.
//
// Resolution rule: check groups sorted by number-of-members descending (largest
// group first).  The first group whose Members are ALL present in
// todo.Priorities wins.
func ResolveColor(todo *model.Todo, cfg config.Config) string {
	if len(todo.Priorities) == 0 || len(cfg.PriorityGroups) == 0 {
		return ""
	}

	// Build a set of the todo's priorities for fast lookup.
	has := make(map[string]bool, len(todo.Priorities))
	for _, p := range todo.Priorities {
		has[p] = true
	}

	// Collect groups and sort largest-first, then alphabetically for stability.
	type namedGroup struct {
		name  string
		group config.PriorityGroup
	}
	groups := make([]namedGroup, 0, len(cfg.PriorityGroups))
	for k, g := range cfg.PriorityGroups {
		groups = append(groups, namedGroup{k, g})
	}
	sort.Slice(groups, func(i, j int) bool {
		li, lj := len(groups[i].group.Members), len(groups[j].group.Members)
		if li != lj {
			return li > lj // larger group first
		}
		return groups[i].name < groups[j].name
	})

	for _, ng := range groups {
		if len(ng.group.Members) == 0 {
			continue
		}
		allMatch := true
		for _, m := range ng.group.Members {
			if !has[m] {
				allMatch = false
				break
			}
		}
		if allMatch {
			return ng.group.Color
		}
	}

	return ""
}
