package sorter_test

import (
	"math"
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
	"github.com/karimStekelenburg/dooing-tmux/internal/sorter"
)

func defaultCfg() config.Config {
	return config.Defaults()
}

func TestGetPriorityScore(t *testing.T) {
	cfg := defaultCfg()

	tests := []struct {
		name string
		todo *model.Todo
		want float64
	}{
		{
			"done todo scores 0",
			&model.Todo{Done: true, Priorities: []string{"important"}},
			0,
		},
		{
			"no priorities scores 0",
			&model.Todo{},
			0,
		},
		{
			"important only, no hours",
			&model.Todo{Priorities: []string{"important"}},
			4.0,
		},
		{
			"urgent only, no hours",
			&model.Todo{Priorities: []string{"urgent"}},
			2.0,
		},
		{
			"both priorities, no hours",
			&model.Todo{Priorities: []string{"important", "urgent"}},
			6.0,
		},
		{
			"important with 1 hour estimate",
			&model.Todo{Priorities: []string{"important"}, EstimatedHours: 1.0},
			4.0 * (1.0 / (1.0 * 0.125)), // 4 * 8 = 32
		},
		{
			"important with 2 hour estimate",
			&model.Todo{Priorities: []string{"important"}, EstimatedHours: 2.0},
			4.0 * (1.0 / (2.0 * 0.125)), // 4 * 4 = 16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sorter.GetPriorityScore(tt.todo, cfg)
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("GetPriorityScore() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestSortWithPriorities(t *testing.T) {
	cfg := defaultCfg()

	high := &model.Todo{ID: "high", Priorities: []string{"important", "urgent"}, CreatedAt: 3}
	medium := &model.Todo{ID: "med", Priorities: []string{"important"}, CreatedAt: 2}
	none := &model.Todo{ID: "none", CreatedAt: 1}

	todos := []*model.Todo{none, medium, high}
	sorter.Sort(todos, false, cfg)

	if todos[0].ID != "high" || todos[1].ID != "med" || todos[2].ID != "none" {
		t.Errorf("expected [high, med, none], got [%s, %s, %s]", todos[0].ID, todos[1].ID, todos[2].ID)
	}
}

func TestShortHighPriorityTaskFirst(t *testing.T) {
	cfg := defaultCfg()

	// Short task with important: score = 4 * (1/(0.5*0.125)) = 4*16 = 64
	short := &model.Todo{ID: "short", Priorities: []string{"important"}, EstimatedHours: 0.5, CreatedAt: 2}
	// Long task with important: score = 4 * (1/(4*0.125)) = 4*2 = 8
	long := &model.Todo{ID: "long", Priorities: []string{"important"}, EstimatedHours: 4.0, CreatedAt: 1}

	todos := []*model.Todo{long, short}
	sorter.Sort(todos, false, cfg)

	if todos[0].ID != "short" {
		t.Errorf("expected short high-priority task first, got %s", todos[0].ID)
	}
}
