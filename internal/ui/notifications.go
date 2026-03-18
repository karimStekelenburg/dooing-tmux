package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

var (
	notifBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("214")).
				Padding(0, 2).
				Width(55)

	notifTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214"))

	notifSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196"))

	notifDueTodaySectionStyle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("214"))

	notifCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	notifFooterStyle = lipgloss.NewStyle().
				Faint(true)
)

// notifItem wraps a todo with its category (overdue or due-today).
type notifItem struct {
	todo     *model.Todo
	overdue  bool // false = due today
}

// notifState holds all state for the due notifications overlay.
type notifState struct {
	open   bool
	items  []notifItem
	cursor int
}

// newNotifState scans todos and returns a notifState (open only if there are items).
func newNotifState(todos []*model.Todo) notifState {
	var items []notifItem
	for _, t := range todos {
		if t.IsOverdue() {
			items = append(items, notifItem{todo: t, overdue: true})
		} else if t.IsDueToday() {
			items = append(items, notifItem{todo: t, overdue: false})
		}
	}
	return notifState{
		open:  len(items) > 0,
		items: items,
	}
}

// jumpToTodoMsg is a tea.Msg that requests the main model to jump its cursor
// to the todo with the given ID.
type jumpToTodoMsg struct {
	todoID string
}

// updateNotifications handles input when the notification overlay is open.
func (m Model) updateNotifications(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "j", "down":
		if m.notif.cursor < len(m.notif.items)-1 {
			m.notif.cursor++
		}
	case "k", "up":
		if m.notif.cursor > 0 {
			m.notif.cursor--
		}
	case "enter":
		if m.notif.cursor >= 0 && m.notif.cursor < len(m.notif.items) {
			id := m.notif.items[m.notif.cursor].todo.ID
			m.notif.open = false
			// Jump cursor to the selected todo in the main list.
			return m, func() tea.Msg { return jumpToTodoMsg{todoID: id} }
		}
	case "q", "esc":
		m.notif.open = false
	}
	return m, nil
}

// renderNotifications returns the styled notification overlay string.
func (m Model) renderNotifications() string {
	// Build dynamic title.
	var overdue, dueToday int
	for _, item := range m.notif.items {
		if item.overdue {
			overdue++
		} else {
			dueToday++
		}
	}
	parts := []string{}
	if overdue > 0 {
		parts = append(parts, fmt.Sprintf("%d overdue", overdue))
	}
	if dueToday > 0 {
		parts = append(parts, fmt.Sprintf("%d due today", dueToday))
	}
	title := " " + strings.Join(parts, ", ") + " "

	var sb strings.Builder
	sb.WriteString(notifTitleStyle.Render(title))
	sb.WriteString("\n\n")

	// Render overdue section.
	if overdue > 0 {
		sb.WriteString(notifSectionStyle.Render("Overdue"))
		sb.WriteString("\n")
		idx := 0
		for _, item := range m.notif.items {
			if !item.overdue {
				continue
			}
			line := renderTodo(item.todo, m.cfg.PriorityGroups)
			if idx == m.notif.cursor {
				sb.WriteString(notifCursorStyle.Render("> ") + line)
			} else {
				sb.WriteString("  " + line)
			}
			sb.WriteString("\n")
			idx++
		}
	}

	// Render due today section.
	if dueToday > 0 {
		if overdue > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(notifDueTodaySectionStyle.Render("Due Today"))
		sb.WriteString("\n")
		idx := overdue
		for _, item := range m.notif.items {
			if item.overdue {
				continue
			}
			line := renderTodo(item.todo, m.cfg.PriorityGroups)
			if idx == m.notif.cursor {
				sb.WriteString(notifCursorStyle.Render("> ") + line)
			} else {
				sb.WriteString("  " + line)
			}
			sb.WriteString("\n")
			idx++
		}
	}

	sb.WriteString("\n")
	sb.WriteString(notifFooterStyle.Render("[j/k] navigate  [Enter] jump to todo  [q/Esc] close"))

	return notifBorderStyle.Render(sb.String())
}
