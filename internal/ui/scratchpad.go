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

const notesIcon = "󱞁"

var (
	scratchpadBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Padding(0, 1)

	scratchpadTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	notesIconStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))
)

// openScratchpad opens the notes editor for the currently selected todo.
func (m *Model) openScratchpad() {
	visible := m.visibleTodos()
	if len(visible) == 0 {
		return
	}
	t := visible[m.cursor]

	ta := textarea.New()
	ta.SetValue(t.Notes)
	ta.ShowLineNumbers = false
	ta.CharLimit = 0 // unlimited
	// Size will be set in View based on terminal dimensions; default to safe values.
	ta.SetWidth(58)
	ta.SetHeight(12)
	ta.Focus()

	m.scratchpad = scratchpadState{
		open:   true,
		todoID: t.ID,
		ta:     ta,
	}
}

// updateScratchpad handles input when the scratchpad overlay is active.
func (m Model) updateScratchpad(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if ok && key.String() == "esc" {
		// Save and close.
		notes := m.scratchpad.ta.Value()
		for _, t := range m.todos {
			if t.ID == m.scratchpad.todoID {
				t.Notes = notes
				break
			}
		}
		_ = m.st.Save(m.storePath, m.todos)
		m.scratchpad = scratchpadState{}
		return m, nil
	}

	// Forward all other input to textarea.
	var cmd tea.Cmd
	m.scratchpad.ta, cmd = m.scratchpad.ta.Update(msg)
	return m, cmd
}

// renderScratchpad returns the styled notes editor overlay string.
func (m Model) renderScratchpad() string {
	// Determine dimensions: ~60% of terminal, minimum 40×10.
	w := m.width * 6 / 10
	if w < 40 {
		w = 40
	}
	h := m.height * 6 / 10
	if h < 10 {
		h = 10
	}

	// Resize textarea to fit inside the border (+padding).
	innerW := w - 4 // account for border + padding
	innerH := h - 5 // account for title lines + footer + border
	if innerH < 4 {
		innerH = 4
	}
	m.scratchpad.ta.SetWidth(innerW)
	m.scratchpad.ta.SetHeight(innerH)

	// Find the todo text for the title.
	todoText := ""
	for _, t := range m.todos {
		if t.ID == m.scratchpad.todoID {
			todoText = t.Text
			break
		}
	}
	maxTitle := w - 12
	if len(todoText) > maxTitle && maxTitle > 0 {
		todoText = todoText[:maxTitle] + "…"
	}

	var sb strings.Builder
	sb.WriteString(scratchpadTitleStyle.Render(" Notes: " + todoText + " "))
	sb.WriteString("\n")
	sb.WriteString(m.scratchpad.ta.View())
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("esc  save & close"))

	return scratchpadBorderStyle.Width(w).Render(sb.String())
}

// renderNotesIcon returns the styled notes icon if the todo has notes.
func renderNotesIcon(notes string) string {
	if notes == "" {
		return ""
	}
	return notesIconStyle.Render(notesIcon)
}
