package ui

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/karimStekelenburg/dooing-tmux/internal/config"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
	"github.com/karimStekelenburg/dooing-tmux/internal/store"
)

// ---- styles ----------------------------------------------------------------

var (
	pendingStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("cyan"))
	inProgressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("yellow"))
	doneStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Faint(true)
	tagStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("green"))
	overdueStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("red")).Bold(true)
	tsStyle         = lipgloss.NewStyle().Faint(true)
	cursorStyle     = lipgloss.NewStyle().Reverse(true)
	titleStyle      = lipgloss.NewStyle().Bold(true)
	footerStyle     = lipgloss.NewStyle().Faint(true)
	borderStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62"))
)

// ---- icons -----------------------------------------------------------------

const (
	iconPending    = "○"
	iconInProgress = "◐"
	iconDone       = "✓"
	iconNotes      = "󱞁"
)

// ---- model -----------------------------------------------------------------

// Model is the root Bubble Tea model.
type Model struct {
	cfg         config.Config
	store       *store.Store
	dataPath    string
	projectMode bool
	todos       []*model.Todo
	cursor      int
	viewport    int // index of the first visible todo
	width       int
	height      int
}

// NewModel creates a new root model, loads config and todos from disk.
func NewModel(projectMode bool) Model {
	cfg, _ := config.Load(config.DefaultConfigPath())
	s := store.New()
	dataPath := store.DefaultPath()
	todos, _ := s.Load(dataPath)

	return Model{
		cfg:         cfg,
		store:       s,
		dataPath:    dataPath,
		projectMode: projectMode,
		todos:       todos,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.todos)-1 {
			m.cursor++
			m.clampViewport()
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.clampViewport()
		}
	case "g":
		m.cursor = 0
		m.viewport = 0
	case "G":
		if len(m.todos) > 0 {
			m.cursor = len(m.todos) - 1
			m.clampViewport()
		}
	}
	return m, nil
}

// clampViewport adjusts the viewport so the cursor remains visible.
func (m *Model) clampViewport() {
	visibleRows := m.visibleRows()
	if m.cursor < m.viewport {
		m.viewport = m.cursor
	} else if m.cursor >= m.viewport+visibleRows {
		m.viewport = m.cursor - visibleRows + 1
	}
	if m.viewport < 0 {
		m.viewport = 0
	}
}

// visibleRows returns how many todo rows fit in the window content area.
func (m *Model) visibleRows() int {
	// Window height minus: 2 border rows + 1 title + 1 blank + 1 footer + 1 blank
	rows := m.cfg.Window.Height - 6
	if rows < 1 {
		rows = 1
	}
	return rows
}

// View implements tea.Model.
func (m Model) View() string {
	title := " Global to-dos "
	if m.projectMode {
		title = " Project to-dos "
	}

	winW := m.cfg.Window.Width
	if m.width > 0 && m.width < winW+4 {
		winW = m.width - 4
	}

	var sb strings.Builder

	// Title line
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	// Body
	if len(m.todos) == 0 {
		sb.WriteString("No todos yet. Press i to create one.")
	} else {
		sb.WriteString(m.renderList(winW))
	}

	// Footer
	sb.WriteString("\n\n")
	sb.WriteString(footerStyle.Render(" [?] for help "))

	return borderStyle.Width(winW).Render(sb.String())
}

// renderList renders the visible slice of todos.
func (m *Model) renderList(width int) string {
	visibleRows := m.visibleRows()
	end := m.viewport + visibleRows
	if end > len(m.todos) {
		end = len(m.todos)
	}

	var lines []string
	for i := m.viewport; i < end; i++ {
		line := m.renderTodo(m.todos[i], width)
		if i == m.cursor {
			line = cursorStyle.Render(line)
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// renderTodo formats a single todo line.
func (m *Model) renderTodo(t *model.Todo, width int) string {
	// Icon
	var icon string
	switch t.GetState() {
	case model.StatePending:
		icon = pendingStyle.Render(iconPending)
	case model.StateInProgress:
		icon = inProgressStyle.Render(iconInProgress)
	case model.StateDone:
		icon = doneStyle.Render(iconDone)
	}

	// Notes indicator
	notesIcon := ""
	if t.Notes != "" {
		notesIcon = " " + iconNotes
	}

	// Colorize #tags in text
	text := colorTags(t.Text, t.GetState() == model.StateDone)

	// Due date
	due := ""
	if t.DueAt != nil {
		d := time.Unix(*t.DueAt, 0)
		if t.IsOverdue() {
			due = " " + overdueStyle.Render(fmt.Sprintf("[! %s]", d.Format("Jan 02, 2006")))
		} else {
			due = " " + fmt.Sprintf("[%s]", d.Format("Jan 02, 2006"))
		}
	}

	// Relative timestamp (right-aligned)
	rel := tsStyle.Render(relativeTime(t.CreatedAt))

	left := fmt.Sprintf("  %s%s %s%s", icon, notesIcon, text, due)
	leftLen := visibleLen(left)
	relLen := visibleLen(rel)
	spaces := width - leftLen - relLen - 2 // -2 for padding
	if spaces < 1 {
		spaces = 1
	}
	return left + strings.Repeat(" ", spaces) + rel
}

// colorTags replaces #tags in text with green-styled versions.
func colorTags(text string, done bool) string {
	words := strings.Fields(text)
	for i, w := range words {
		if strings.HasPrefix(w, "#") {
			if done {
				words[i] = doneStyle.Render(w)
			} else {
				words[i] = tagStyle.Render(w)
			}
		} else if done {
			words[i] = doneStyle.Render(w)
		}
	}
	return strings.Join(words, " ")
}

// relativeTime returns a human-readable relative time string.
func relativeTime(unixSec int64) string {
	d := time.Since(time.Unix(unixSec, 0))
	switch {
	case d < time.Minute:
		return "@just now"
	case d < time.Hour:
		return fmt.Sprintf("@%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("@%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("@%dd ago", int(d.Hours()/24))
	}
}

// visibleLen returns the display width of a string, ignoring ANSI escape codes.
func visibleLen(s string) int {
	// Strip ANSI sequences with a simple state machine.
	inEscape := false
	count := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		count++
	}
	return count
}
