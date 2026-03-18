package ui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

var (
	foldedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)

// nestedState holds all state for nested-task related features.
type nestedState struct {
	// folded is the set of todo IDs whose children are hidden.
	folded map[string]bool
}

func newNestedState() nestedState {
	return nestedState{folded: make(map[string]bool)}
}

// isFolded returns true if the todo with the given ID has folded children.
func (s *nestedState) isFolded(id string) bool {
	return s.folded[id]
}

// toggleFold toggles the fold state of the given todo ID.
func (s *nestedState) toggleFold(id string) {
	s.folded[id] = !s.folded[id]
}

// hasChildren returns true if any todo in the list has parentID as its parent.
func hasChildren(todos []*model.Todo, parentID string) bool {
	for _, t := range todos {
		if t.ParentID == parentID {
			return true
		}
	}
	return false
}

// visibleTodos returns the filtered and fold-aware slice of todos for display.
// Todos whose ancestor is folded are omitted.
func (m Model) visibleTodos() []*model.Todo {
	filtered := m.filteredTodos()

	// Build a set of IDs that are hidden because an ancestor is folded.
	hidden := make(map[string]bool)
	for _, t := range filtered {
		if m.nested.isFolded(t.ID) {
			// Mark all children (and their children) hidden.
			markChildrenHidden(filtered, t.ID, hidden)
		}
	}

	var result []*model.Todo
	for _, t := range filtered {
		if !hidden[t.ID] {
			result = append(result, t)
		}
	}
	return result
}

// markChildrenHidden recursively marks all todos with the given parentID as hidden.
func markChildrenHidden(todos []*model.Todo, parentID string, hidden map[string]bool) {
	for _, t := range todos {
		if t.ParentID == parentID {
			hidden[t.ID] = true
			markChildrenHidden(todos, t.ID, hidden)
		}
	}
}

// createNestedTodo opens the text input to create a child of the current todo.
func (m *Model) createNestedTodo() tea.Cmd {
	visible := m.visibleTodos()
	if len(visible) == 0 {
		return nil
	}
	parent := visible[m.cursor]
	m.nestedParentID = parent.ID
	m.nestedParentDepth = parent.Depth
	m.inputMode = inputModeCreateNested
	m.ti.SetValue("")
	m.ti.Placeholder = "Child todo… (#tag to categorise)"
	m.ti.Focus()
	return textinput.Blink
}

// insertionIndexForChild returns the index in m.todos where a new child of parentID should be inserted.
// It is placed after the parent and all its existing descendants (depth-first).
func (m Model) insertionIndexForChild(parentID string) int {
	parentIdx := -1
	for i, t := range m.todos {
		if t.ID == parentID {
			parentIdx = i
			break
		}
	}
	if parentIdx < 0 {
		return len(m.todos)
	}

	// Walk forward collecting all descendants (depth-first).
	end := parentIdx + 1
	for end < len(m.todos) {
		if isDescendant(m.todos, m.todos[end].ID, parentID) {
			end++
		} else {
			break
		}
	}
	return end
}

// isDescendant returns true if todoID has parentID as an ancestor (any depth).
func isDescendant(todos []*model.Todo, todoID, ancestorID string) bool {
	// Build parent map.
	parentMap := make(map[string]string, len(todos))
	for _, t := range todos {
		parentMap[t.ID] = t.ParentID
	}
	cur := todoID
	for {
		p, ok := parentMap[cur]
		if !ok || p == "" {
			return false
		}
		if p == ancestorID {
			return true
		}
		cur = p
	}
}

// promoteOrphans promotes todos whose parent no longer exists to depth=0, parentID="".
func promoteOrphans(todos []*model.Todo) {
	idSet := make(map[string]bool, len(todos))
	for _, t := range todos {
		idSet[t.ID] = true
	}
	for _, t := range todos {
		if t.ParentID != "" && !idSet[t.ParentID] {
			t.ParentID = ""
			t.Depth = 0
		}
	}
}

// countDescendants returns the total number of todos that are descendants of parentID.
func countDescendants(todos []*model.Todo, parentID string) int {
	count := 0
	for _, t := range todos {
		if t.ParentID == parentID {
			count++
			count += countDescendants(todos, t.ID)
		}
	}
	return count
}

// itoa converts an int to string.
func itoa(n int) string {
	return strconv.Itoa(n)
}

// renderIndent returns the indentation prefix for a todo at the given depth.
// Each depth level adds indentSize spaces.
func renderIndent(depth, indentSize int) string {
	if depth <= 0 || indentSize <= 0 {
		return ""
	}
	return strings.Repeat(" ", depth*indentSize)
}
