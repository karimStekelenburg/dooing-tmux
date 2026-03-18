package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
	"github.com/karimStekelenburg/dooing-tmux/internal/store"
)

// inputMode describes what the text input is currently doing.
type inputMode int

const (
	inputModeNone   inputMode = iota
	inputModeCreate           // 'i' — new todo
	inputModeEdit             // 'e' — edit existing todo
)

// undoEntry stores a deleted todo for possible restoration.
type undoEntry struct {
	todo          *model.Todo
	originalIndex int
}

const maxUndoStack = 100

var (
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62"))

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	doneStyle = lipgloss.NewStyle().
			Faint(true).
			Strikethrough(true)

	inProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))

	pendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	inputBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("212")).
				Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Italic(true)

	dialogBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("196")).
				Padding(0, 2)

	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("196"))

	dialogFooterStyle = lipgloss.NewStyle().
				Faint(true)
)

// Model is the root Bubble Tea model for dooing-tmux.
type Model struct {
	projectMode bool
	width       int
	height      int

	todos     []*model.Todo
	cursor    int
	storePath string
	st        *store.Store

	inputMode inputMode
	editingID string // set when inputMode == inputModeEdit
	ti        textinput.Model

	// Confirmation dialog state.
	showConfirm    bool
	confirmTodoIdx int // index of todo pending delete confirmation

	// Undo stack (in-memory only, not persisted).
	undoStack []undoEntry

	statusMsg string // transient flash message
}

// NewModel creates a new root model, loading todos from disk.
func NewModel(projectMode bool) Model {
	path := store.DefaultPath()

	st := store.New()
	todos, _ := st.Load(path)

	ti := textinput.New()
	ti.Placeholder = "Type your todo… (#tag to categorise)"
	ti.CharLimit = 500
	ti.Width = 50

	return Model{
		projectMode: projectMode,
		todos:       todos,
		storePath:   path,
		st:          st,
		ti:          ti,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Always handle window resize.
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		return m, nil
	}

	// Confirmation dialog blocks all other input.
	if m.showConfirm {
		return m.updateConfirm(msg)
	}

	// If text input is active, route keys there.
	if m.inputMode != inputModeNone {
		return m.updateInput(msg)
	}

	return m.updateNormal(msg)
}

func (m Model) updateConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "y", "Y":
		m = m.deleteTodoAt(m.confirmTodoIdx)
		m.showConfirm = false
		m.confirmTodoIdx = 0
	case "n", "N", "q", "esc":
		m.showConfirm = false
		m.confirmTodoIdx = 0
	}
	return m, nil
}

func (m Model) updateInput(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.ti, cmd = m.ti.Update(msg)
		return m, cmd
	}

	switch key.String() {
	case "enter":
		text := strings.TrimSpace(m.ti.Value())
		if text == "" {
			return m, nil
		}
		switch m.inputMode {
		case inputModeCreate:
			t := model.NewTodo(text)
			m.todos = append(m.todos, t)
			m.cursor = len(m.todos) - 1
			_ = m.st.Save(m.storePath, m.todos)
		case inputModeEdit:
			for _, t := range m.todos {
				if t.ID == m.editingID {
					t.Text = text
					t.Category = t.ExtractCategory()
					break
				}
			}
			_ = m.st.Save(m.storePath, m.todos)
		}
		m.inputMode = inputModeNone
		m.editingID = ""
		m.ti.SetValue("")
		m.ti.Blur()
		return m, nil

	case "esc":
		m.inputMode = inputModeNone
		m.editingID = ""
		m.ti.SetValue("")
		m.ti.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.ti, cmd = m.ti.Update(msg)
	return m, cmd
}

func (m Model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// Clear status message on any keypress.
	m.statusMsg = ""

	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	// Navigation
	case "j", "down":
		if m.cursor < len(m.todos)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}

	// Create
	case "i":
		m.inputMode = inputModeCreate
		m.ti.SetValue("")
		m.ti.Placeholder = "Type your todo… (#tag to categorise)"
		m.ti.Focus()
		return m, textinput.Blink

	// Edit
	case "e":
		if len(m.todos) == 0 {
			break
		}
		t := m.todos[m.cursor]
		m.inputMode = inputModeEdit
		m.editingID = t.ID
		m.ti.SetValue(t.Text)
		m.ti.CursorEnd()
		m.ti.Focus()
		return m, textinput.Blink

	// Toggle
	case "x":
		if len(m.todos) == 0 {
			break
		}
		m.todos[m.cursor].Toggle()
		_ = m.st.Save(m.storePath, m.todos)

	// Delete
	case "d":
		if len(m.todos) == 0 {
			break
		}
		t := m.todos[m.cursor]
		if t.GetState() == model.StateDone {
			// Done todo: delete immediately.
			m = m.deleteTodoAt(m.cursor)
		} else {
			// Incomplete: ask for confirmation.
			m.showConfirm = true
			m.confirmTodoIdx = m.cursor
		}

	// Delete all completed
	case "D":
		var remaining []*model.Todo
		for i, t := range m.todos {
			if t.GetState() == model.StateDone {
				m.pushUndo(t, i)
			} else {
				remaining = append(remaining, t)
			}
		}
		if remaining == nil {
			remaining = []*model.Todo{}
		}
		m.todos = remaining
		if m.cursor >= len(m.todos) && len(m.todos) > 0 {
			m.cursor = len(m.todos) - 1
		} else if len(m.todos) == 0 {
			m.cursor = 0
		}
		_ = m.st.Save(m.storePath, m.todos)

	// Undo
	case "u":
		if len(m.undoStack) == 0 {
			break
		}
		last := m.undoStack[len(m.undoStack)-1]
		m.undoStack = m.undoStack[:len(m.undoStack)-1]

		idx := last.originalIndex
		if idx > len(m.todos) {
			idx = len(m.todos)
		}

		// Re-insert at original index.
		m.todos = append(m.todos, nil)
		copy(m.todos[idx+1:], m.todos[idx:])
		m.todos[idx] = last.todo
		m.cursor = idx

		_ = m.st.Save(m.storePath, m.todos)
		m.statusMsg = "Todo restored"
	}

	return m, nil
}

// deleteTodoAt removes the todo at idx, saves to disk, and adjusts cursor.
// It also pushes an undo entry.
func (m Model) deleteTodoAt(idx int) Model {
	if idx < 0 || idx >= len(m.todos) {
		return m
	}
	t := m.todos[idx]
	m.pushUndo(t, idx)
	m.todos = append(m.todos[:idx], m.todos[idx+1:]...)
	if m.cursor >= len(m.todos) && len(m.todos) > 0 {
		m.cursor = len(m.todos) - 1
	} else if len(m.todos) == 0 {
		m.cursor = 0
	}
	_ = m.st.Save(m.storePath, m.todos)
	return m
}

// pushUndo adds an entry to the undo stack, capping at maxUndoStack.
func (m *Model) pushUndo(t *model.Todo, idx int) {
	// Deep copy to avoid mutation after deletion.
	cp := *t
	entry := undoEntry{todo: &cp, originalIndex: idx}
	m.undoStack = append(m.undoStack, entry)
	if len(m.undoStack) > maxUndoStack {
		m.undoStack = m.undoStack[len(m.undoStack)-maxUndoStack:]
	}
}

// View implements tea.Model.
func (m Model) View() string {
	title := " Global to-dos "
	if m.projectMode {
		title = " Project to-dos "
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	if len(m.todos) == 0 && m.inputMode == inputModeNone {
		sb.WriteString(lipgloss.NewStyle().Faint(true).Render("No todos yet. Press i to create one."))
		sb.WriteString("\n")
	}

	for i, t := range m.todos {
		line := renderTodo(t)
		if i == m.cursor && m.inputMode == inputModeNone && !m.showConfirm {
			line = cursorStyle.Render("> ") + line
		} else {
			line = "  " + line
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Inline input rendering
	if m.inputMode != inputModeNone {
		sb.WriteString("\n")
		prompt := "New todo:"
		if m.inputMode == inputModeEdit {
			prompt = "Edit todo:"
		}
		inputBlock := inputBorderStyle.Render(
			lipgloss.NewStyle().Bold(true).Render(prompt) + "\n" + m.ti.View(),
		)
		sb.WriteString(inputBlock)
		sb.WriteString("\n")
	}

	// Confirmation dialog
	if m.showConfirm && m.confirmTodoIdx < len(m.todos) {
		todoText := m.todos[m.confirmTodoIdx].Text
		// Truncate for display.
		if len(todoText) > 50 {
			todoText = todoText[:47] + "…"
		}
		dialogContent := dialogTitleStyle.Render(" Delete incomplete todo? ") + "\n\n" +
			lipgloss.NewStyle().Faint(true).Render(todoText) + "\n\n" +
			dialogFooterStyle.Render(" [Y]es - [N]o ")
		sb.WriteString("\n")
		sb.WriteString(dialogBorderStyle.Render(dialogContent))
		sb.WriteString("\n")
	}

	// Status / flash message
	if m.statusMsg != "" {
		sb.WriteString("\n")
		sb.WriteString(statusStyle.Render(m.statusMsg))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("[?] for help"))

	return borderStyle.Render(sb.String())
}

// renderTodo returns a styled single-line representation of t.
func renderTodo(t *model.Todo) string {
	var icon string
	var textStyle lipgloss.Style

	switch t.GetState() {
	case model.StateDone:
		icon = "✓"
		textStyle = doneStyle
	case model.StateInProgress:
		icon = "◐"
		textStyle = inProgressStyle
	default:
		icon = "○"
		textStyle = pendingStyle
	}

	displayText := highlightTags(t.Text, textStyle)
	return fmt.Sprintf("%s %s", lipgloss.NewStyle().Bold(true).Render(icon), displayText)
}

// highlightTags renders the todo text with #tags coloured, applying baseStyle to non-tag parts.
func highlightTags(text string, baseStyle lipgloss.Style) string {
	parts := tagRegexForUI.Split(text, -1)
	tags := tagRegexForUI.FindAllString(text, -1)

	var sb strings.Builder
	for i, part := range parts {
		sb.WriteString(baseStyle.Render(part))
		if i < len(tags) {
			sb.WriteString(tagStyle.Render(tags[i]))
		}
	}
	return sb.String()
}
