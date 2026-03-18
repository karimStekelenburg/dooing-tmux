package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// searchState holds all state for the search popup.
type searchState struct {
	open    bool
	input   textinput.Model
	results []*model.Todo
	cursor  int
}

var (
	searchBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("212")).
				Padding(0, 1).
				Width(40)

	searchTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	searchCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("212"))

	searchResultStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	searchNoMatchStyle = lipgloss.NewStyle().
				Faint(true)

	searchFooterStyle = lipgloss.NewStyle().
				Faint(true)
)

// newSearchState creates a new search state with a ready text input.
func newSearchState() searchState {
	ti := textinput.New()
	ti.Placeholder = "Search todos…"
	ti.CharLimit = 200
	ti.Width = 36
	return searchState{input: ti}
}

// openSearch opens the search popup.
func (m *Model) openSearch() tea.Cmd {
	m.search.open = true
	m.search.cursor = 0
	m.search.results = nil
	m.search.input.SetValue("")
	m.search.input.Focus()
	return textinput.Blink
}

// runSearch filters todos by the current search query.
func (m *Model) runSearch() {
	query := strings.ToLower(strings.TrimSpace(m.search.input.Value()))
	if query == "" {
		m.search.results = nil
		m.search.cursor = 0
		return
	}

	var results []*model.Todo
	for _, t := range m.todos {
		if strings.Contains(strings.ToLower(t.Text), query) {
			results = append(results, t)
		}
	}
	m.search.results = results
	m.search.cursor = 0
}

// updateSearch handles input when the search popup is open.
func (m Model) updateSearch(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		// Route to text input.
		var cmd tea.Cmd
		m.search.input, cmd = m.search.input.Update(msg)
		m.runSearch()
		return m, cmd
	}

	switch key.String() {
	case "q", "esc":
		m.search.open = false
		m.search.input.Blur()

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
			selected := m.search.results[m.search.cursor]
			// Find index in m.todos and jump cursor there.
			for i, t := range m.todos {
				if t.ID == selected.ID {
					m.cursor = i
					break
				}
			}
			m.search.open = false
			m.search.input.Blur()
		}

	default:
		// Route other keys to text input (typing).
		var cmd tea.Cmd
		m.search.input, cmd = m.search.input.Update(key)
		m.runSearch()
		return m, cmd
	}

	return m, nil
}

// renderSearch returns the styled search popup string.
func (m Model) renderSearch() string {
	var sb strings.Builder
	sb.WriteString(searchTitleStyle.Render(" Search "))
	sb.WriteString("\n\n")
	sb.WriteString(m.search.input.View())
	sb.WriteString("\n\n")

	if m.search.input.Value() == "" {
		sb.WriteString(searchNoMatchStyle.Render("Type to search…"))
	} else if len(m.search.results) == 0 {
		sb.WriteString(searchNoMatchStyle.Render("No matches found"))
	} else {
		for i, t := range m.search.results {
			// Truncate long todo text.
			text := t.Text
			if len(text) > 32 {
				text = text[:29] + "…"
			}
			icon := todoIcon(t)
			var line string
			if i == m.search.cursor {
				line = searchCursorStyle.Render("> "+icon+" ") + searchResultStyle.Render(text)
			} else {
				line = "  " + icon + " " + searchResultStyle.Render(text)
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(searchFooterStyle.Render("[j/k] navigate  [enter] jump  [q] close"))

	return searchBorderStyle.Render(sb.String())
}

// todoIcon returns the status icon string for a todo.
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
