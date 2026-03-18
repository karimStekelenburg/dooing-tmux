package priority

import (
	"math"
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

func defaultCfg() config.Config {
	return config.Defaults()
}

func newTodo(text string, priorities []string, hours float64, done bool) *model.Todo {
	t := model.NewTodo(text)
	t.Priorities = priorities
	t.EstimatedHours = hours
	t.Done = done
	return t
}

// ---- GetScore tests ----

func TestDoneTodoScoresZero(t *testing.T) {
	todo := newTodo("done task", []string{"important", "urgent"}, 0, true)
	score := GetScore(todo, defaultCfg())
	if score != 0 {
		t.Errorf("expected 0 for done todo, got %f", score)
	}
}

func TestNoPrioritiesScoresZero(t *testing.T) {
	todo := newTodo("plain task", nil, 0, false)
	score := GetScore(todo, defaultCfg())
	if score != 0 {
		t.Errorf("expected 0 for todo with no priorities, got %f", score)
	}
}

func TestBaseSinglePriority(t *testing.T) {
	// "important" weight = 4; no hours → multiplier = 1.0
	todo := newTodo("task", []string{"important"}, 0, false)
	score := GetScore(todo, defaultCfg())
	if score != 4.0 {
		t.Errorf("expected 4.0, got %f", score)
	}
}

func TestBaseMultiplePriorities(t *testing.T) {
	// important(4) + urgent(2) = 6; no hours → 6.0
	todo := newTodo("task", []string{"important", "urgent"}, 0, false)
	score := GetScore(todo, defaultCfg())
	if score != 6.0 {
		t.Errorf("expected 6.0, got %f", score)
	}
}

func TestScoreWithEstimatedHours(t *testing.T) {
	// base = 4 (important only); hours=2, hourScoreValue=0.125
	// multiplier = 1 / (2 * 0.125) = 1 / 0.25 = 4.0
	// score = 4 * 4 = 16.0
	todo := newTodo("task", []string{"important"}, 2, false)
	score := GetScore(todo, defaultCfg())
	if math.Abs(score-16.0) > 1e-9 {
		t.Errorf("expected 16.0, got %f", score)
	}
}

func TestScoreHigherForShorterTask(t *testing.T) {
	// Both have same priorities; shorter task should score higher.
	short := newTodo("short", []string{"important"}, 1, false)
	long := newTodo("long", []string{"important"}, 8, false)
	sScore := GetScore(short, defaultCfg())
	lScore := GetScore(long, defaultCfg())
	if sScore <= lScore {
		t.Errorf("short task (score %f) should score higher than long (score %f)", sScore, lScore)
	}
}

func TestUnknownPriorityNamesIgnored(t *testing.T) {
	todo := newTodo("task", []string{"nonexistent"}, 0, false)
	score := GetScore(todo, defaultCfg())
	if score != 0 {
		t.Errorf("expected 0 for unknown priority names, got %f", score)
	}
}

// ---- ResolveColor tests ----

func TestNoGroupsReturnsEmpty(t *testing.T) {
	cfg := config.Config{PriorityGroups: nil}
	todo := newTodo("task", []string{"important"}, 0, false)
	color := ResolveColor(todo, cfg)
	if color != "" {
		t.Errorf("expected empty color, got %q", color)
	}
}

func TestNoPrioritiesReturnsEmpty(t *testing.T) {
	todo := newTodo("task", nil, 0, false)
	color := ResolveColor(todo, defaultCfg())
	if color != "" {
		t.Errorf("expected empty color for todo with no priorities, got %q", color)
	}
}

func TestHighGroupBothImportantUrgent(t *testing.T) {
	// "high" group requires both "important" AND "urgent".
	todo := newTodo("task", []string{"important", "urgent"}, 0, false)
	color := ResolveColor(todo, defaultCfg())
	// high group color is "#ff0000"
	if color != "#ff0000" {
		t.Errorf("expected #ff0000 (high group), got %q", color)
	}
}

func TestMediumGroupImportantOnly(t *testing.T) {
	// "medium" group requires only "important"
	todo := newTodo("task", []string{"important"}, 0, false)
	color := ResolveColor(todo, defaultCfg())
	// medium group color is "#ffff00"
	if color != "#ffff00" {
		t.Errorf("expected #ffff00 (medium group), got %q", color)
	}
}

func TestLowGroupUrgentOnly(t *testing.T) {
	// "low" group requires only "urgent"
	todo := newTodo("task", []string{"urgent"}, 0, false)
	color := ResolveColor(todo, defaultCfg())
	// low group color is "#0000ff"
	if color != "#0000ff" {
		t.Errorf("expected #0000ff (low group), got %q", color)
	}
}

func TestNoMatchingGroupReturnsEmpty(t *testing.T) {
	todo := newTodo("task", []string{"unknown_priority"}, 0, false)
	color := ResolveColor(todo, defaultCfg())
	if color != "" {
		t.Errorf("expected empty color for unmatched priorities, got %q", color)
	}
}
