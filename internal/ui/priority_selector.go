package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
)

// prioritySelectorState holds the state for the priority selector overlay.
type prioritySelectorState struct {
	open     bool
	cursor   int
	todoID   string   // ID of the todo being edited
	names    []string // priority names from config
	selected []bool   // checkbox state per priority
}

var (
	priorityBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("214")).
				Padding(0, 1).
				Width(38)

	priorityTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	priorityFooterStyle = lipgloss.NewStyle().
				Faint(true)

	priorityItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	priorityCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))
)

// openPrioritySelector initialises the priority selector for the todo at the current cursor.
func (m *Model) openPrioritySelector() {
	visible := m.filteredTodos()
	if len(visible) == 0 {
		return
	}
	t := visible[m.cursor]

	names := make([]string, len(m.cfg.Priorities))
	selected := make([]bool, len(m.cfg.Priorities))
	currentSet := make(map[string]bool, len(t.Priorities))
	for _, p := range t.Priorities {
		currentSet[p] = true
	}
	for i, p := range m.cfg.Priorities {
		names[i] = p.Name
		selected[i] = currentSet[p.Name]
	}

	m.prioritySel = prioritySelectorState{
		open:     true,
		cursor:   0,
		todoID:   t.ID,
		names:    names,
		selected: selected,
	}
}

// updatePrioritySelector handles input when the priority selector is open.
func (m Model) updatePrioritySelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "q", "esc":
		m.prioritySel.open = false

	case "j", "down":
		if m.prioritySel.cursor < len(m.prioritySel.names)-1 {
			m.prioritySel.cursor++
		}

	case "k", "up":
		if m.prioritySel.cursor > 0 {
			m.prioritySel.cursor--
		}

	case " ":
		if m.prioritySel.cursor < len(m.prioritySel.selected) {
			m.prioritySel.selected[m.prioritySel.cursor] = !m.prioritySel.selected[m.prioritySel.cursor]
		}

	case "enter":
		m.applyPrioritySelection()
		m.prioritySel.open = false
	}

	return m, nil
}

// applyPrioritySelection writes the selected priorities back to the todo.
func (m *Model) applyPrioritySelection() {
	var priorities []string
	for i, sel := range m.prioritySel.selected {
		if sel {
			priorities = append(priorities, m.prioritySel.names[i])
		}
	}

	for _, t := range m.todos {
		if t.ID == m.prioritySel.todoID {
			t.Priorities = priorities
			break
		}
	}
	m.sortTodos()
	_ = m.st.Save(m.storePath, m.todos)
}

// renderPrioritySelector returns the styled priority selector overlay.
func (m Model) renderPrioritySelector() string {
	var sb strings.Builder
	sb.WriteString(priorityTitleStyle.Render(" Priorities "))
	sb.WriteString("\n\n")

	for i, name := range m.prioritySel.names {
		checkbox := "[ ]"
		if m.prioritySel.selected[i] {
			checkbox = "[x]"
		}

		var line string
		if i == m.prioritySel.cursor {
			line = priorityCursorStyle.Render("> "+checkbox+" ") + priorityItemStyle.Render(name)
		} else {
			line = "  " + priorityItemStyle.Render(checkbox+" "+name)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(priorityFooterStyle.Render("[space] toggle  [enter] confirm  [q] cancel"))

	return priorityBorderStyle.Render(sb.String())
}

// resolvePriorityColor returns the color for a todo based on priority group matching.
// Groups are checked largest-members-first; the first group whose ALL members are
// present in the todo's priorities wins.
func resolvePriorityColor(priorities []string, groups map[string]config.PriorityGroup) string {
	if len(priorities) == 0 || len(groups) == 0 {
		return ""
	}

	pSet := make(map[string]bool, len(priorities))
	for _, p := range priorities {
		pSet[p] = true
	}

	// Find the group with the most members that fully matches.
	bestColor := ""
	bestSize := 0
	for _, g := range groups {
		if len(g.Members) <= bestSize {
			continue
		}
		allPresent := true
		for _, m := range g.Members {
			if !pSet[m] {
				allPresent = false
				break
			}
		}
		if allPresent {
			bestColor = g.Color
			bestSize = len(g.Members)
		}
	}

	return bestColor
}

// renderPriorityLabel returns a styled string like "[important, urgent]" for display.
func renderPriorityLabel(priorities []string, groups map[string]config.PriorityGroup) string {
	if len(priorities) == 0 {
		return ""
	}
	label := fmt.Sprintf("[%s]", strings.Join(priorities, ", "))
	color := resolvePriorityColor(priorities, groups)
	if color != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(label)
	}
	return lipgloss.NewStyle().Faint(true).Render(label)
}
