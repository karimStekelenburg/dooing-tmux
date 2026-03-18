package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// timeInputState holds the state for the time estimation input overlay.
type timeInputState struct {
	open   bool
	todoID string
	ti     textinput.Model
	errMsg string
}

var (
	timeInputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("214")).
				Padding(0, 1)

	timeInputTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	timeInputErrStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))

	timeEstStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
)

// openTimeInput opens the time estimation input for the currently selected todo.
func (m *Model) openTimeInput() {
	visible := m.visibleTodos()
	if len(visible) == 0 {
		return
	}
	t := visible[m.cursor]
	ti := textinput.New()
	ti.Placeholder = "e.g. 30m, 2h, 1d, 0.5w"
	ti.CharLimit = 20
	ti.Width = 30
	// Pre-fill with existing estimate if present.
	if t.EstimatedHours > 0 {
		ti.SetValue(formatHoursToInput(t.EstimatedHours))
	}
	ti.Focus()
	m.timeInput = timeInputState{
		open:   true,
		todoID: t.ID,
		ti:     ti,
	}
}

// updateTimeInput handles input when the time estimation overlay is active.
func (m Model) updateTimeInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.timeInput.ti, cmd = m.timeInput.ti.Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "esc", "q":
		m.timeInput = timeInputState{}
	case "enter":
		raw := strings.TrimSpace(m.timeInput.ti.Value())
		if raw == "" {
			m.timeInput = timeInputState{}
			return m, nil
		}
		hours, err := parseTimeInput(raw)
		if err != nil {
			m.timeInput.errMsg = err.Error()
			m.timeInput.ti.Focus()
			return m, textinput.Blink
		}
		// Apply to todo.
		for _, t := range m.todos {
			if t.ID == m.timeInput.todoID {
				t.EstimatedHours = hours
				break
			}
		}
		m.sortTodos()
		_ = m.st.Save(m.storePath, m.todos)
		m.timeInput = timeInputState{}
		m.statusMsg = fmt.Sprintf("Time estimate set: %s", formatHours(hours))
	default:
		var cmd tea.Cmd
		m.timeInput.ti, cmd = m.timeInput.ti.Update(msg)
		return m, cmd
	}

	return m, nil
}

// renderTimeInput returns the styled time estimation input overlay.
func (m Model) renderTimeInput() string {
	var sb strings.Builder
	sb.WriteString(timeInputTitleStyle.Render(" Set time estimate "))
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Units: m(inutes) h(ours) d(ays=8h) w(eeks=40h)"))
	sb.WriteString("\n")
	sb.WriteString(m.timeInput.ti.View())
	if m.timeInput.errMsg != "" {
		sb.WriteString("\n")
		sb.WriteString(timeInputErrStyle.Render(m.timeInput.errMsg))
	}
	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("enter confirm  esc cancel"))
	return timeInputBorderStyle.Render(sb.String())
}

// parseTimeInput parses a time string like "30m", "2h", "1d", "0.5w" into hours.
// Returns an error if the format is not recognised.
func parseTimeInput(s string) (float64, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid format: use e.g. 30m, 2h, 1d, 0.5w")
	}
	unit := s[len(s)-1]
	numStr := s[:len(s)-1]
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil || val <= 0 {
		return 0, fmt.Errorf("invalid number %q — use e.g. 30m, 2h, 1d, 0.5w", numStr)
	}
	switch unit {
	case 'm':
		return val / 60.0, nil
	case 'h':
		return val, nil
	case 'd':
		return val * 8.0, nil
	case 'w':
		return val * 40.0, nil
	default:
		return 0, fmt.Errorf("unknown unit %q — use m, h, d, or w", string(unit))
	}
}

// formatHours converts a float64 hours value to the best-fit human-readable string.
// < 1h  → minutes  (e.g. "30m")
// < 8h  → hours    (e.g. "2h")
// < 40h → days     (e.g. "1d")
// >= 40h → weeks   (e.g. "0.5w")
func formatHours(hours float64) string {
	switch {
	case hours < 1.0:
		mins := hours * 60.0
		return formatFloat(mins) + "m"
	case hours < 8.0:
		return formatFloat(hours) + "h"
	case hours < 40.0:
		days := hours / 8.0
		return formatFloat(days) + "d"
	default:
		weeks := hours / 40.0
		return formatFloat(weeks) + "w"
	}
}

// formatHoursToInput converts stored hours back to an input-friendly string.
func formatHoursToInput(hours float64) string {
	return formatHours(hours)
}

// formatFloat formats a float, stripping trailing zeros.
func formatFloat(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	// Trim trailing zeros after decimal point.
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

// renderTimeEstimate returns the styled [≈ Xunit] label for display in todo lines.
// Returns "" if no estimate.
func renderTimeEstimate(hours float64) string {
	if hours <= 0 {
		return ""
	}
	label := fmt.Sprintf("[≈ %s]", formatHours(hours))
	return timeEstStyle.Render(label)
}
