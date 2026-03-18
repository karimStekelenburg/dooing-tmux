package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62"))
)

// Model is the root Bubble Tea model for dooing-tmux.
type Model struct {
	projectMode bool
	width       int
	height      int
}

// NewModel creates a new root model.
func NewModel(projectMode bool) Model {
	return Model{projectMode: projectMode}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	title := " Global to-dos "
	if m.projectMode {
		title = " Project to-dos "
	}

	body := titleStyle.Render(title) + "\n\n" +
		"No todos yet. Press i to create one.\n\n" +
		lipgloss.NewStyle().Faint(true).Render("[?] for help")

	return borderStyle.Render(body)
}
