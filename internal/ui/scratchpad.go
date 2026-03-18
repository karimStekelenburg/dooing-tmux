package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// scratchpadState holds the state for the per-todo notes editor overlay.
type scratchpadState struct {
	open   bool
	todoID string
	ta     textarea.Model
}

var (
	scratchpadBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("62")).
				Padding(0, 1)

	scratchpadTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("62"))

	scratchpadFooterStyle = lipgloss.NewStyle().
				Faint(true)

	notesIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62"))
)

// notesIcon is the icon shown on todos that have notes.
const notesIcon = "󱞁"

// newScratchpadState creates a new scratchpad state with a ready textarea.
func newScratchpadState() scratchpadState {
	ta := textarea.New()
	ta.Placeholder = "Write your notes here…"
	ta.ShowLineNumbers = false
	ta.CharLimit = 10000
	return scratchpadState{ta: ta}
}

// openScratchpad opens the notes editor for the selected todo.
func (m *Model) openScratchpad() tea.Cmd {
	visible := m.filteredTodos()
	if len(visible) == 0 {
		return nil
	}
	t := visible[m.cursor]

	// Size the textarea to ~60% of terminal (or minimum 40x10).
	w := m.width * 60 / 100
	h := m.height * 60 / 100
	if w < 40 {
		w = 40
	}
	if h < 10 {
		h = 10
	}

	m.pad.todoID = t.ID
	m.pad.ta.SetWidth(w)
	m.pad.ta.SetHeight(h)
	m.pad.ta.SetValue(t.Notes)
	m.pad.ta.CursorEnd()
	m.pad.open = true

	return m.pad.ta.Focus()
}

// updateScratchpad handles input when the scratchpad overlay is open.
func (m Model) updateScratchpad(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		// Route to textarea.
		var cmd tea.Cmd
		m.pad.ta, cmd = m.pad.ta.Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "esc":
		// Save notes and close.
		notes := m.pad.ta.Value()
		for _, t := range m.todos {
			if t.ID == m.pad.todoID {
				t.Notes = notes
				break
			}
		}
		_ = m.st.Save(m.storePath, m.todos)
		m.pad.open = false
		m.pad.ta.Blur()
		return m, nil

	default:
		// Route all other keys to textarea.
		var cmd tea.Cmd
		m.pad.ta, cmd = m.pad.ta.Update(key)
		return m, cmd
	}
}

// renderScratchpad returns the styled scratchpad overlay string.
func (m Model) renderScratchpad() string {
	var sb strings.Builder

	// Title: truncate todo text if needed.
	todoText := ""
	for _, t := range m.todos {
		if t.ID == m.pad.todoID {
			todoText = t.Text
			break
		}
	}
	if len(todoText) > 40 {
		todoText = todoText[:37] + "…"
	}

	sb.WriteString(scratchpadTitleStyle.Render(" Notes: " + todoText + " "))
	sb.WriteString("\n\n")
	sb.WriteString(m.pad.ta.View())
	sb.WriteString("\n\n")
	sb.WriteString(scratchpadFooterStyle.Render("[esc] save & close"))

	return scratchpadBorderStyle.Render(sb.String())
}
