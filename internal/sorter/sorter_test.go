package sorter_test

import (
	"testing"
	"time"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
	"github.com/karimStekelenburg/dooing-tmux/internal/sorter"
)

func makeTime(offset int64) int64 {
	return time.Now().Unix() + offset
}


func TestUndoneBeforeDone(t *testing.T) {
	done := &model.Todo{ID: "1", Text: "done", Done: true, CreatedAt: 1}
	pending := &model.Todo{ID: "2", Text: "pending", CreatedAt: 2}

	todos := []*model.Todo{done, pending}
	sorter.Sort(todos, false)

	if todos[0].ID != "2" {
		t.Errorf("expected pending first, got %q", todos[0].Text)
	}
}

func TestDoneSortByCompletedTime(t *testing.T) {
	earlier := makeTime(-100)
	later := makeTime(-10)

	a := &model.Todo{ID: "a", Done: true, CompletedAt: &earlier, CreatedAt: 1}
	b := &model.Todo{ID: "b", Done: true, CompletedAt: &later, CreatedAt: 2}

	todos := []*model.Todo{a, b}
	sorter.Sort(todos, true)

	if todos[0].ID != "b" {
		t.Errorf("expected more-recently completed first, got %q", todos[0].ID)
	}
}

func TestDoneSortByCompletedTimeDisabled(t *testing.T) {
	earlier := makeTime(-100)
	later := makeTime(-10)

	a := &model.Todo{ID: "a", Done: true, CompletedAt: &earlier, CreatedAt: 1}
	b := &model.Todo{ID: "b", Done: true, CompletedAt: &later, CreatedAt: 2}

	todos := []*model.Todo{a, b}
	sorter.Sort(todos, false)

	// When doneSortByCompleted=false, falls through to creation time.
	if todos[0].ID != "a" {
		t.Errorf("expected creation-time order when doneSortByCompleted=false, got %q", todos[0].ID)
	}
}

func TestDueDateEarlierFirst(t *testing.T) {
	dueEarly := makeTime(100)
	dueLate := makeTime(500)

	a := &model.Todo{ID: "a", DueAt: &dueLate, CreatedAt: 1}
	b := &model.Todo{ID: "b", DueAt: &dueEarly, CreatedAt: 2}

	todos := []*model.Todo{a, b}
	sorter.Sort(todos, false)

	if todos[0].ID != "b" {
		t.Errorf("expected earlier due date first, got %q", todos[0].ID)
	}
}

func TestWithDueDateBeforeWithout(t *testing.T) {
	due := makeTime(200)
	hasDue := &model.Todo{ID: "a", DueAt: &due, CreatedAt: 1}
	noDue := &model.Todo{ID: "b", CreatedAt: 2}

	todos := []*model.Todo{noDue, hasDue}
	sorter.Sort(todos, false)

	if todos[0].ID != "a" {
		t.Errorf("expected todo with due date first, got %q", todos[0].ID)
	}
}

func TestCreationTimeAscending(t *testing.T) {
	a := &model.Todo{ID: "a", CreatedAt: 10}
	b := &model.Todo{ID: "b", CreatedAt: 5}

	todos := []*model.Todo{a, b}
	sorter.Sort(todos, false)

	if todos[0].ID != "b" {
		t.Errorf("expected earlier creation time first, got %q", todos[0].ID)
	}
}

func TestStableSort(t *testing.T) {
	// Two todos with identical sort keys — relative order must be preserved.
	now := int64(1000)
	a := &model.Todo{ID: "a", CreatedAt: now}
	b := &model.Todo{ID: "b", CreatedAt: now}

	todos := []*model.Todo{a, b}
	sorter.Sort(todos, false)

	// a was first, b was second; with equal keys they must stay that way.
	if todos[0].ID != "a" || todos[1].ID != "b" {
		t.Errorf("stable sort violated: got [%s, %s]", todos[0].ID, todos[1].ID)
	}
}

func TestInProgressTreatedAsUndone(t *testing.T) {
	done := &model.Todo{ID: "done", Done: true, CreatedAt: 1}
	inProgress := &model.Todo{ID: "ip", InProgress: true, CreatedAt: 2}

	todos := []*model.Todo{done, inProgress}
	sorter.Sort(todos, false)

	if todos[0].ID != "ip" {
		t.Errorf("expected in-progress before done, got %q", todos[0].ID)
	}
}

func TestSortNestedPreservesTreeOrder(t *testing.T) {
	root1 := &model.Todo{ID: "r1", Text: "root1", CreatedAt: 1}
	root2 := &model.Todo{ID: "r2", Text: "root2", CreatedAt: 2}
	child1a := &model.Todo{ID: "c1a", Text: "child1a", ParentID: "r1", Depth: 1, CreatedAt: 3}
	child1b := &model.Todo{ID: "c1b", Text: "child1b", ParentID: "r1", Depth: 1, CreatedAt: 4}
	grandchild := &model.Todo{ID: "gc", Text: "grandchild", ParentID: "c1a", Depth: 2, CreatedAt: 5}

	todos := []*model.Todo{grandchild, root2, child1b, root1, child1a}
	result := sorter.SortNested(todos, false)

	// Expected order: r1, c1a, gc, c1b, r2
	expected := []string{"r1", "c1a", "gc", "c1b", "r2"}
	for i, id := range expected {
		if result[i].ID != id {
			t.Errorf("position %d: expected %q got %q", i, id, result[i].ID)
		}
	}
}

func TestSortNestedParentDeletion(t *testing.T) {
	// After parent deletion and orphan promotion, children should be top-level.
	orphan := &model.Todo{ID: "o1", Text: "orphan", ParentID: "", Depth: 0, CreatedAt: 1}
	other := &model.Todo{ID: "o2", Text: "other", ParentID: "", Depth: 0, CreatedAt: 2}

	todos := []*model.Todo{other, orphan}
	result := sorter.SortNested(todos, false)
	if len(result) != 2 {
		t.Fatalf("expected 2 todos, got %d", len(result))
	}
}
