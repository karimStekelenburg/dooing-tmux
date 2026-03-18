package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
	"github.com/karimStekelenburg/dooing-tmux/internal/model"
	"github.com/karimStekelenburg/dooing-tmux/internal/sorter"
	"github.com/karimStekelenburg/dooing-tmux/internal/store"
)

// inputMode describes what the text input is currently doing.
type inputMode int

const (
	inputModeNone         inputMode = iota
	inputModeCreate                 // 'i' — new todo
	inputModeEdit                   // 'e' — edit existing todo
	inputModeCreateNested           // 'n' — new child todo
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

	helpBorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Width(50)

	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("62"))

	helpSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("214"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	helpDescStyle = lipgloss.NewStyle().
			Faint(true)

	quickKeysBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("240")).
				Padding(0, 1)
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

	cfg config.Config

	inputMode inputMode
	editingID string // set when inputMode == inputModeEdit
	ti        textinput.Model

	// Confirmation dialog state.
	showConfirm    bool
	confirmTodoIdx int // index of todo pending delete confirmation

	// Undo stack (in-memory only, not persisted).
	undoStack []undoEntry

	statusMsg string // transient flash message

	// Help window state.
	showHelp bool

	// Tag window state.
	tagWin      tagWindowState
	activeFilter string // currently active tag filter (empty = no filter)

	// Priority selector state.
	prioritySel prioritySelectorState

	// Nested tasks state.
	nested            nestedState
	nestedParentID    string // set when inputMode == inputModeCreateNested
	nestedParentDepth int    // depth of the parent todo

	// Calendar popup state.
	cal calendarState

	// Time estimation input state.
	timeInput timeInputState

	// Search overlay state.
	search searchState

	// Scratchpad (notes editor) state.
	scratchpad scratchpadState

	// Project mode state.
	projectDirName string // basename of git root for title; empty = global
	projectErr     string // set when project mode failed (not in git repo, etc.)
}

// NewModel creates a new root model, loading todos from disk.
func NewModel(projectMode bool) Model {
	cfg, _ := config.Load(config.DefaultConfigPath())

	path := store.DefaultPath()
	var projectDirName string
	var projectErr string

	if projectMode {
		projPath, err := projectStorePath(cfg.ProjectFile)
		if err != nil {
			projectErr = err.Error()
		} else {
			path = projPath
			// Extract dirname for the title.
			root := filepath.Dir(projPath)
			projectDirName = filepath.Base(root)
			// Handle gitignore.
			ensureGitignore(root, cfg.ProjectFile, cfg.AutoGitignore)
		}
	}

	st := store.New()
	todos, _ := st.Load(path)

	// Sort on load so initial display is correct.
	todos = sorter.SortNested(todos, cfg.DoneSortByCompleted, cfg)

	ti := textinput.New()
	ti.Placeholder = "Type your todo… (#tag to categorise)"
	ti.CharLimit = 500
	ti.Width = 50

	return Model{
		projectMode:    projectMode,
		projectDirName: projectDirName,
		projectErr:     projectErr,
		todos:          todos,
		storePath:      path,
		st:             st,
		cfg:            cfg,
		ti:             ti,
		tagWin:         newTagWindowState(),
		nested:         newNestedState(),
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

	// Help window intercepts input when open.
	if m.showHelp {
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.String() == "q" || key.String() == "?" {
				m.showHelp = false
			}
		}
		return m, nil
	}

	// Tag window intercepts input when open.
	if m.tagWin.open {
		return m.updateTagWindow(msg)
	}

	// Priority selector intercepts input when open.
	if m.prioritySel.open {
		return m.updatePrioritySelector(msg)
	}

	// Calendar intercepts input when open.
	if m.cal.open {
		return m.updateCalendar(msg)
	}

	// Time estimation input intercepts when open.
	if m.timeInput.open {
		return m.updateTimeInput(msg)
	}

	// Search overlay intercepts when open.
	if m.search.open {
		return m.updateSearch(msg)
	}

	// Scratchpad intercepts when open.
	if m.scratchpad.open {
		return m.updateScratchpad(msg)
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
			m.cursor = len(m.todos) - 1 // point at new todo before sort
			m.sortTodos()               // cursor follows the new todo by ID
			_ = m.st.Save(m.storePath, m.todos)
		case inputModeCreateNested:
			t := model.NewTodo(text)
			t.ParentID = m.nestedParentID
			t.Depth = m.nestedParentDepth + 1
			// Insert after parent and all existing children.
			insertAt := m.insertionIndexForChild(m.nestedParentID)
			m.todos = append(m.todos, nil)
			copy(m.todos[insertAt+1:], m.todos[insertAt:])
			m.todos[insertAt] = t
			m.cursor = insertAt
			m.sortTodos()
			_ = m.st.Save(m.storePath, m.todos)
			m.nestedParentID = ""
			m.nestedParentDepth = 0
		case inputModeEdit:
			for _, t := range m.todos {
				if t.ID == m.editingID {
					t.Text = text
					t.Category = t.ExtractCategory()
					break
				}
			}
			m.sortTodos()
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
		m.nestedParentID = ""
		m.nestedParentDepth = 0
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

	// Build visible (fold+filter aware) slice once.
	visible := m.visibleTodos()

	switch key.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	// Navigation
	case "j", "down":
		if m.cursor < len(visible)-1 {
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

	// Create nested child
	case "n":
		return m, m.createNestedTodo()

	// Fold/unfold
	case "z", "tab":
		if len(visible) > 0 {
			t := visible[m.cursor]
			if hasChildren(m.todos, t.ID) {
				m.nested.toggleFold(t.ID)
				// Keep cursor in bounds after fold.
				newVisible := m.visibleTodos()
				if m.cursor >= len(newVisible) && len(newVisible) > 0 {
					m.cursor = len(newVisible) - 1
				}
			}
		}

	// Edit
	case "e":
		if len(visible) == 0 {
			break
		}
		t := visible[m.cursor]
		m.inputMode = inputModeEdit
		m.editingID = t.ID
		m.ti.SetValue(t.Text)
		m.ti.CursorEnd()
		m.ti.Focus()
		return m, textinput.Blink

	// Tags window
	case "t":
		m.tagWin.open = true
		m.tagWin.cursor = 0
		m.tagWin.mode = tagWindowBrowse
		m.tagWin.refreshTags(m.todos)

	// Priority selector
	case "p":
		m.openPrioritySelector()

	// Calendar (due date)
	case "H":
		m.openCalendar()

	// Time estimate
	case "T":
		m.openTimeInput()

	// Remove time estimate
	case "R":
		if len(visible) > 0 {
			t := visible[m.cursor]
			for _, todo := range m.todos {
				if todo.ID == t.ID {
					todo.EstimatedHours = 0
					break
				}
			}
			m.sortTodos()
			_ = m.st.Save(m.storePath, m.todos)
			m.statusMsg = "Time estimate removed"
		}

	// Reload from disk
	case "f":
		fresh, err := m.st.Load(m.storePath)
		if err == nil {
			m.todos = sorter.SortNested(fresh, m.cfg.DoneSortByCompleted, m.cfg)
			vis := m.visibleTodos()
			if m.cursor >= len(vis) && len(vis) > 0 {
				m.cursor = len(vis) - 1
			}
			m.statusMsg = "Reloaded from disk"
		} else {
			m.statusMsg = "Reload failed: " + err.Error()
		}

	// Search
	case "/":
		m.openSearch()

	// Scratchpad / notes
	case "s":
		m.openScratchpad()

	// Remove due date
	case "r":
		if len(visible) > 0 {
			m.removeDueDate()
		}

	// Clear filter
	case "c":
		m.activeFilter = ""
		m.cursor = 0

	// Help
	case "?":
		m.showHelp = true

	// Toggle
	case "x":
		if len(visible) == 0 {
			break
		}
		visible[m.cursor].Toggle()
		m.sortTodos()
		_ = m.st.Save(m.storePath, m.todos)

	// Delete
	case "d":
		if len(visible) == 0 {
			break
		}
		t := visible[m.cursor]
		todosIdx := m.findTodosIndex(t.ID)
		if t.GetState() == model.StateDone {
			// Done todo: delete immediately.
			m = m.deleteTodoAt(todosIdx)
		} else {
			// Incomplete: ask for confirmation.
			m.showConfirm = true
			m.confirmTodoIdx = todosIdx
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
		promoteOrphans(m.todos)
		m.sortTodos()
		vis := m.visibleTodos()
		if m.cursor >= len(vis) && len(vis) > 0 {
			m.cursor = len(vis) - 1
		} else if len(vis) == 0 {
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
		m.cursor = idx // point at restored todo before sort
		m.sortTodos()  // cursor follows the restored todo by ID

		_ = m.st.Save(m.storePath, m.todos)
		m.statusMsg = "Todo restored"
	}

	return m, nil
}

// deleteTodoAt removes the todo at idx (index into m.todos), saves to disk, and adjusts cursor.
// It also pushes an undo entry.
func (m Model) deleteTodoAt(idx int) Model {
	if idx < 0 || idx >= len(m.todos) {
		return m
	}
	t := m.todos[idx]
	m.pushUndo(t, idx)
	m.todos = append(m.todos[:idx], m.todos[idx+1:]...)
	// Promote children whose parent was just deleted.
	promoteOrphans(m.todos)
	vis := m.visibleTodos()
	if m.cursor >= len(vis) && len(vis) > 0 {
		m.cursor = len(vis) - 1
	} else if len(vis) == 0 {
		m.cursor = 0
	}
	_ = m.st.Save(m.storePath, m.todos)
	return m
}

// findTodosIndex returns the index of the todo with the given ID in m.todos, or -1.
func (m Model) findTodosIndex(id string) int {
	for i, t := range m.todos {
		if t.ID == id {
			return i
		}
	}
	return -1
}

// sortTodos sorts todos in-place and updates cursor to follow the previously selected todo.
func (m *Model) sortTodos() {
	if len(m.todos) == 0 {
		return
	}
	// Remember which todo was under the cursor.
	var selectedID string
	if m.cursor >= 0 && m.cursor < len(m.todos) {
		selectedID = m.todos[m.cursor].ID
	}

	m.todos = sorter.SortNested(m.todos, m.cfg.DoneSortByCompleted, m.cfg)

	// Re-locate the cursor.
	if selectedID != "" {
		for i, t := range m.todos {
			if t.ID == selectedID {
				m.cursor = i
				return
			}
		}
	}
	// Fallback: clamp cursor.
	if m.cursor >= len(m.todos) {
		m.cursor = len(m.todos) - 1
	}
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
		if m.projectDirName != "" {
			title = " " + m.projectDirName + " to-dos "
		} else {
			title = " Project to-dos "
		}
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(title))
	sb.WriteString("\n\n")

	// Show project error prominently if set.
	if m.projectErr != "" {
		errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
		sb.WriteString(errStyle.Render("Project error: " + m.projectErr))
		sb.WriteString("\n\n")
	}

	// Filter header (2 lines when active).
	if header := m.renderFilterHeader(); header != "" {
		sb.WriteString(header)
	}

	visible := m.visibleTodos()
	if len(visible) == 0 && m.inputMode == inputModeNone {
		if m.activeFilter != "" {
			sb.WriteString(lipgloss.NewStyle().Faint(true).Render("No todos matching #" + m.activeFilter + "."))
		} else {
			sb.WriteString(lipgloss.NewStyle().Faint(true).Render("No todos yet. Press i to create one."))
		}
		sb.WriteString("\n")
	}

	indentSize := m.cfg.IndentSize
	if indentSize <= 0 {
		indentSize = 2
	}

	for i, t := range visible {
		line := renderTodo(t, m.cfg.PriorityGroups)
		// Add fold indicator if this todo has folded children.
		if m.nested.isFolded(t.ID) && hasChildren(m.todos, t.ID) {
			childCount := countDescendants(m.todos, t.ID)
			line += " " + foldedStyle.Render("[+"+itoa(childCount)+"]")
		}
		indent := renderIndent(t.Depth, indentSize)
		if i == m.cursor && m.inputMode == inputModeNone && !m.showConfirm {
			line = cursorStyle.Render("> ") + indent + line
		} else {
			line = "  " + indent + line
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Inline input rendering
	if m.inputMode != inputModeNone {
		sb.WriteString("\n")
		prompt := "New todo:"
		switch m.inputMode {
		case inputModeEdit:
			prompt = "Edit todo:"
		case inputModeCreateNested:
			prompt = "New child todo:"
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

	mainView := borderStyle.Render(sb.String())

	// Tag window overlay — rendered left of main.
	if m.tagWin.open {
		tagView := m.renderTagWindow()
		return lipgloss.JoinHorizontal(lipgloss.Top, tagView, "  ", mainView)
	}

	// Priority selector overlay — rendered left of main.
	if m.prioritySel.open {
		priView := m.renderPrioritySelector()
		return lipgloss.JoinHorizontal(lipgloss.Top, priView, "  ", mainView)
	}

	// Calendar overlay — rendered left of main.
	if m.cal.open {
		calView := m.renderCalendar()
		return lipgloss.JoinHorizontal(lipgloss.Top, calView, "  ", mainView)
	}

	// Time estimation input — rendered left of main.
	if m.timeInput.open {
		tiView := m.renderTimeInput()
		return lipgloss.JoinHorizontal(lipgloss.Top, tiView, "  ", mainView)
	}

	// Search overlay — rendered left of main.
	if m.search.open {
		searchView := m.renderSearch()
		return lipgloss.JoinHorizontal(lipgloss.Top, searchView, "  ", mainView)
	}

	// Scratchpad overlay — rendered left of main.
	if m.scratchpad.open {
		padView := m.renderScratchpad()
		return lipgloss.JoinHorizontal(lipgloss.Top, padView, "  ", mainView)
	}

	// Help window overlay — rendered side-by-side (right of main).
	if m.showHelp {
		help := renderHelpWindow()
		return lipgloss.JoinHorizontal(lipgloss.Top, mainView, "  ", help)
	}

	// Quick keys panel below main window (optional, config-driven).
	if m.cfg.QuickKeys {
		qk := renderQuickKeys()
		return lipgloss.JoinVertical(lipgloss.Left, mainView, qk)
	}

	return mainView
}

// renderHelpWindow returns the styled help overlay string.
func renderHelpWindow() string {
	type binding struct {
		key  string
		desc string
	}
	type section struct {
		title    string
		bindings []binding
	}

	sections := []section{
		{
			title: "Main window",
			bindings: []binding{
				{"i", "Create new todo"},
				{"n", "Create child todo (nested)"},
				{"z / tab", "Fold/unfold children"},
				{"e", "Edit selected todo"},
				{"H", "Set due date (calendar popup)"},
				{"r", "Remove due date"},
				{"T", "Set time estimate (30m, 2h, 1d, 0.5w)"},
				{"R", "Remove time estimate"},
				{"/", "Open search"},
				{"s", "Open notes/scratchpad (Esc saves)"},
				{"x", "Toggle todo status (pending → in progress → done)"},
				{"d", "Delete selected todo"},
				{"D", "Delete all completed todos"},
				{"u", "Undo last deletion"},
				{"j / ↓", "Move cursor down"},
				{"k / ↑", "Move cursor up"},
				{"f", "Reload todos from disk"},
				{"?", "Toggle this help window"},
				{"q / ctrl+c", "Quit"},
			},
		},
		{
			title: "Tags window (open with t)",
			bindings: []binding{
				{"t", "Open tag window"},
				{"enter", "Filter by selected tag"},
				{"e", "Rename selected tag"},
				{"d", "Delete selected tag from all todos"},
				{"c", "Clear active filter (main window)"},
				{"q / esc", "Close tag window"},
			},
		},
		{
			title: "Priority selector (open with p)",
			bindings: []binding{
				{"p", "Open priority selector"},
				{"space", "Toggle priority checkbox"},
				{"j / k", "Navigate priorities"},
				{"enter", "Confirm selection"},
				{"q / esc", "Cancel"},
			},
		},
	}

	var sb strings.Builder
	sb.WriteString(helpTitleStyle.Render(" Keybindings "))

	for _, sec := range sections {
		sb.WriteString("\n\n")
		sb.WriteString(helpSectionStyle.Render(sec.title))
		sb.WriteString("\n")
		for _, b := range sec.bindings {
			sb.WriteString("  ")
			sb.WriteString(helpKeyStyle.Render(fmt.Sprintf("%-20s", b.key)))
			sb.WriteString(helpDescStyle.Render(b.desc))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(lipgloss.NewStyle().Faint(true).Render("Press ? or q to close"))

	return helpBorderStyle.Render(sb.String())
}

// renderQuickKeys returns a small two-column quick reference panel.
func renderQuickKeys() string {
	keys := [][2]string{
		{"i", "create"},
		{"e", "edit"},
		{"x", "toggle"},
		{"d", "delete"},
		{"D", "del done"},
		{"u", "undo"},
		{"t", "tags"},
		{"p", "priority"},
		{"c", "clr filter"},
		{"j/k", "navigate"},
		{"q", "quit"},
		{"?", "help"},
	}

	half := (len(keys) + 1) / 2
	var left, right strings.Builder
	for idx, k := range keys {
		entry := fmt.Sprintf(" %s %s ",
			helpKeyStyle.Render(fmt.Sprintf("%-7s", k[0])),
			helpDescStyle.Render(k[1]),
		)
		if idx < half {
			left.WriteString(entry)
			left.WriteString("\n")
		} else {
			right.WriteString(entry)
			right.WriteString("\n")
		}
	}

	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		left.String(),
		"  ",
		right.String(),
	)
	return quickKeysBorderStyle.Render(cols)
}

// renderTodo returns a styled single-line representation of t.
func renderTodo(t *model.Todo, groups map[string]config.PriorityGroup) string {
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
	line := fmt.Sprintf("%s %s", lipgloss.NewStyle().Bold(true).Render(icon), displayText)

	// Append time estimate if present.
	if est := renderTimeEstimate(t.EstimatedHours); est != "" {
		line += " " + est
	}

	// Append due date if present.
	if t.DueAt != nil {
		line += " " + formatDueDate(t.DueAt, t.Done)
	}

	// Append notes icon if present.
	if icon := renderNotesIcon(t.Notes); icon != "" {
		line += " " + icon
	}

	// Append priority label if present.
	if len(t.Priorities) > 0 {
		line += " " + renderPriorityLabel(t.Priorities, groups)
	}

	// Apply priority group color to the whole line if resolved.
	color := resolvePriorityColor(t.Priorities, groups)
	if color != "" && t.GetState() != model.StateDone {
		line = lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(line)
	}

	return line
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
