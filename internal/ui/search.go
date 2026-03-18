package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// searchState holds the state for the search popup.
type searchState struct {
	open    bool
	ti      textinput.Model
	results []*model.Todo // matching todos (subset of m.todos)
	cursor  int
}

var (
	searchBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1).
				Width(42)

	searchTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("62"))

	searchResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	searchCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	searchEmptyStyle = lipgloss.NewStyle().
				Faint(true)
)

// openSearch opens the search overlay.
func (m *Model) openSearch() {
	ti := textinput.New()
	ti.Placeholder = "Search todos…"
	ti.CharLimit = 200
	ti.Width = 36
	ti.Focus()
	m.search = searchState{
		open: true,
		ti:   ti,
	}
	m.search.results = m.searchTodos("")
}

// updateSearch handles input when the search overlay is active.
func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		// Let textinput handle character input.
		var cmd tea.Cmd
		m.search.ti, cmd = m.search.ti.Update(msg)
		// Live filter.
		m.search.results = m.searchTodos(m.search.ti.Value())
		m.search.cursor = 0
		return m, cmd
	}

	switch key.String() {
	case "esc", "q":
		m.search = searchState{}
	case "j", "down":
		if m.search.cursor < len(m.search.results)-1 {
			m.search.cursor++
		}
	case "k", "up":
		if m.search.cursor > 0 {
			m.search.cursor--
		}
	case "enter":
		if len(m.search.results) > 0 {
			target := m.search.results[m.search.cursor]
			// Jump cursor to this todo in the main list.
			visible := m.visibleTodos()
			for i, t := range visible {
				if t.ID == target.ID {
					m.cursor = i
					break
				}
			}
			m.search = searchState{}
		}
	default:
		var cmd tea.Cmd
		m.search.ti, cmd = m.search.ti.Update(msg)
		// Live filter on any other key.
		m.search.results = m.searchTodos(m.search.ti.Value())
		m.search.cursor = 0
		return m, cmd
	}

	return m, nil
}

// searchTodos returns todos whose text contains query (case-insensitive).
func (m Model) searchTodos(query string) []*model.Todo {
	query = strings.ToLower(query)
	var results []*model.Todo
	for _, t := range m.todos {
		if query == "" || strings.Contains(strings.ToLower(t.Text), query) {
			results = append(results, t)
		}
	}
	return results
}

// renderSearch returns the styled search popup string.
func (m Model) renderSearch() string {
	var sb strings.Builder
	sb.WriteString(searchTitleStyle.Render(" Search "))
	sb.WriteString("\n")
	sb.WriteString(m.search.ti.View())
	sb.WriteString("\n\n")

	if len(m.search.results) == 0 {
		sb.WriteString(searchEmptyStyle.Render("No matches found"))
	} else {
		// Show up to 8 results.
		start := 0
		if m.search.cursor >= 8 {
			start = m.search.cursor - 7
		}
		end := start + 8
		if end > len(m.search.results) {
			end = len(m.search.results)
		}
		for i := start; i < end; i++ {
			t := m.search.results[i]
			icon := todoIcon(t)
			text := t.Text
			if len(text) > 28 {
				text = text[:27] + "…"
			}
			line := fmt.Sprintf("%s %s", icon, text)
			if i == m.search.cursor {
				sb.WriteString(searchCursorStyle.Render("> " + line))
			} else {
				sb.WriteString(searchResultStyle.Render("  " + line))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("j/k navigate  enter jump  esc close"))
	return searchBorderStyle.Render(sb.String())
}

// todoIcon returns the status icon for a todo.
func todoIcon(t *model.Todo) string {
	switch t.GetState() {
	case model.StateDone:
		return "✓"
	case model.StateInProgress:
		return "◐"
	default:
		return "○"
	}
}
