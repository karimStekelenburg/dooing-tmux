package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// prioritySelectorState holds all state for the priority selector overlay.
type prioritySelectorState struct {
	open      bool
	todoID    string   // ID of the todo being edited
	items     []string // priority names from config (ordered)
	checked   []bool   // checked[i] == true means items[i] is selected
	cursor    int
}

var (
	priSelBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("212")).
				Padding(0, 2).
				Width(38)

	priSelTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	priSelItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	priSelCheckedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true)

	priSelCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	priSelFooterStyle = lipgloss.NewStyle().
				Faint(true)
)

// openPrioritySelector populates and opens the priority selector for the todo at visibleIdx.
func (m Model) openPrioritySelector(todo *model.Todo) Model {
	names := make([]string, len(m.cfg.Priorities))
	for i, p := range m.cfg.Priorities {
		names[i] = p.Name
	}

	// Pre-populate checked state.
	has := make(map[string]bool, len(todo.Priorities))
	for _, p := range todo.Priorities {
		has[p] = true
	}
	checked := make([]bool, len(names))
	for i, n := range names {
		checked[i] = has[n]
	}

	m.priSel = prioritySelectorState{
		open:    true,
		todoID:  todo.ID,
		items:   names,
		checked: checked,
		cursor:  0,
	}
	return m
}

// updatePrioritySelector handles all keystrokes when the priority selector is open.
func (m Model) updatePrioritySelector(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "q", "esc":
		m.priSel.open = false

	case "j", "down":
		if m.priSel.cursor < len(m.priSel.items)-1 {
			m.priSel.cursor++
		}

	case "k", "up":
		if m.priSel.cursor > 0 {
			m.priSel.cursor--
		}

	case " ":
		if len(m.priSel.items) > 0 {
			m.priSel.checked[m.priSel.cursor] = !m.priSel.checked[m.priSel.cursor]
		}

	case "enter":
		// Collect selected priority names.
		var selected []string
		for i, name := range m.priSel.items {
			if m.priSel.checked[i] {
				selected = append(selected, name)
			}
		}
		// Update the todo.
		for _, t := range m.todos {
			if t.ID == m.priSel.todoID {
				t.Priorities = selected
				break
			}
		}
		m.sortTodos()
		_ = m.st.Save(m.storePath, m.todos)
		m.priSel.open = false
	}

	return m, nil
}

// renderPrioritySelector returns the styled priority selector overlay.
func (m Model) renderPrioritySelector() string {
	var sb strings.Builder
	sb.WriteString(priSelTitleStyle.Render(" Select Priorities "))
	sb.WriteString("\n\n")

	if len(m.priSel.items) == 0 {
		sb.WriteString(priSelFooterStyle.Render("No priorities configured"))
		sb.WriteString("\n")
	} else {
		for i, name := range m.priSel.items {
			checkbox := "[ ]"
			nameRender := priSelItemStyle.Render(name)
			if m.priSel.checked[i] {
				checkbox = "[x]"
				nameRender = priSelCheckedStyle.Render(name)
			}

			var line string
			if i == m.priSel.cursor {
				line = priSelCursorStyle.Render("> ") + checkbox + " " + nameRender
			} else {
				line = "  " + checkbox + " " + nameRender
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(priSelFooterStyle.Render("[space] toggle  [j/k] navigate  [enter] confirm  [q] cancel"))

	return priSelBorderStyle.Render(sb.String())
}

// priorityLabel returns a short label for the todo's priorities, e.g. "[important,urgent]".
func priorityLabel(todo *model.Todo) string {
	if len(todo.Priorities) == 0 {
		return ""
	}
	return fmt.Sprintf("[%s]", strings.Join(todo.Priorities, ","))
}
