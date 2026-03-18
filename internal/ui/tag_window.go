package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// tagWindowMode describes the sub-state within the tag window.
type tagWindowMode int

const (
	tagWindowBrowse  tagWindowMode = iota
	tagWindowRename                // 'e' pressed — waiting for new name input
)

// tagWindowState holds all state for the tag window overlay.
type tagWindowState struct {
	open         bool
	cursor       int
	tags         []string // sorted unique list; refreshed when window opens / tags change
	mode         tagWindowMode
	renameInput  textinput.Model
	renamingTag  string // tag currently being renamed
}

var (
	tagWindowBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Padding(0, 1).
				Width(28)

	tagWindowTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("86"))

	tagItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	tagCursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))

	activeFilterStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("226")) // yellow

	tagWindowFooterStyle = lipgloss.NewStyle().
				Faint(true)

	tagRenameBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("212")).
				Padding(0, 1)
)

// newTagWindowState initialises a tagWindowState with a ready text input.
func newTagWindowState() tagWindowState {
	ti := textinput.New()
	ti.Placeholder = "new tag name…"
	ti.CharLimit = 64
	ti.Width = 24
	return tagWindowState{renameInput: ti}
}

// refreshTags rebuilds the tag list from the current todos.
func (s *tagWindowState) refreshTags(todos []*model.Todo) {
	s.tags = model.GetAllTags(todos)
	if s.cursor >= len(s.tags) && len(s.tags) > 0 {
		s.cursor = len(s.tags) - 1
	}
	if len(s.tags) == 0 {
		s.cursor = 0
	}
}

// updateTagWindow handles all keystrokes when the tag window is open.
// It returns the (possibly-modified) model and a command.
func (m Model) updateTagWindow(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		if m.tagWin.mode == tagWindowRename {
			var cmd tea.Cmd
			m.tagWin.renameInput, cmd = m.tagWin.renameInput.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	if m.tagWin.mode == tagWindowRename {
		return m.updateTagWindowRename(key)
	}

	return m.updateTagWindowBrowse(key)
}

func (m Model) updateTagWindowBrowse(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "q", "esc":
		m.tagWin.open = false
		m.tagWin.cursor = 0

	case "j", "down":
		if m.tagWin.cursor < len(m.tagWin.tags)-1 {
			m.tagWin.cursor++
		}

	case "k", "up":
		if m.tagWin.cursor > 0 {
			m.tagWin.cursor--
		}

	case "enter":
		// Apply filter.
		if len(m.tagWin.tags) == 0 {
			break
		}
		m.activeFilter = m.tagWin.tags[m.tagWin.cursor]
		m.cursor = 0
		m.tagWin.open = false

	case "e":
		// Start rename.
		if len(m.tagWin.tags) == 0 {
			break
		}
		m.tagWin.mode = tagWindowRename
		m.tagWin.renamingTag = m.tagWin.tags[m.tagWin.cursor]
		m.tagWin.renameInput.SetValue("")
		m.tagWin.renameInput.Placeholder = "new tag name…"
		m.tagWin.renameInput.Focus()
		return m, textinput.Blink

	case "d":
		// Delete tag from all todos.
		if len(m.tagWin.tags) == 0 {
			break
		}
		tag := m.tagWin.tags[m.tagWin.cursor]
		removeTagFromTodos(m.todos, tag)
		_ = m.st.Save(m.storePath, m.todos)
		// Clear filter if we deleted the active filter tag.
		if m.activeFilter == tag {
			m.activeFilter = ""
		}
		m.tagWin.refreshTags(m.todos)
		m.sortTodos()
	}

	return m, nil
}

func (m Model) updateTagWindowRename(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "enter":
		newName := strings.TrimSpace(m.tagWin.renameInput.Value())
		if newName != "" && newName != m.tagWin.renamingTag {
			renameTagInTodos(m.todos, m.tagWin.renamingTag, newName)
			_ = m.st.Save(m.storePath, m.todos)
			// Update filter if we renamed the active filter.
			if m.activeFilter == m.tagWin.renamingTag {
				m.activeFilter = newName
			}
			m.tagWin.refreshTags(m.todos)
		}
		m.tagWin.mode = tagWindowBrowse
		m.tagWin.renamingTag = ""
		m.tagWin.renameInput.Blur()
		m.tagWin.renameInput.SetValue("")

	case "esc":
		m.tagWin.mode = tagWindowBrowse
		m.tagWin.renamingTag = ""
		m.tagWin.renameInput.Blur()
		m.tagWin.renameInput.SetValue("")

	default:
		var cmd tea.Cmd
		m.tagWin.renameInput, cmd = m.tagWin.renameInput.Update(key)
		return m, cmd
	}

	return m, nil
}

// renderTagWindow returns the styled tag window string.
func (m Model) renderTagWindow() string {
	var sb strings.Builder
	sb.WriteString(tagWindowTitleStyle.Render(" Tags "))
	sb.WriteString("\n\n")

	if len(m.tagWin.tags) == 0 {
		sb.WriteString(tagWindowFooterStyle.Render("No tags found"))
		sb.WriteString("\n")
	} else {
		for i, tag := range m.tagWin.tags {
			var line string
			if i == m.tagWin.cursor {
				line = tagCursorStyle.Render("> ") + tagStyle.Render("#"+tag)
			} else {
				line = "  " + tagItemStyle.Render("#"+tag)
			}
			if tag == m.activeFilter {
				line += " " + activeFilterStyle.Render("(active)")
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	if m.tagWin.mode == tagWindowRename {
		sb.WriteString("\n")
		sb.WriteString(tagRenameBorderStyle.Render(
			lipgloss.NewStyle().Bold(true).Render("Rename #"+m.tagWin.renamingTag+":") +
				"\n" + m.tagWin.renameInput.View(),
		))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(tagWindowFooterStyle.Render("[enter] filter  [e] rename  [d] delete  [q] close"))

	return tagWindowBorderStyle.Render(sb.String())
}

// tagBoundaryRegex returns a regex that matches #tag only when not followed by more word chars.
// This prevents #work from matching #workout.
func tagBoundaryRegex(tag string) *regexp.Regexp {
	return regexp.MustCompile(`#` + regexp.QuoteMeta(tag) + `\b`)
}

// removeTagFromTodos removes #tagname (exact, word-boundary) from the text of all todos.
func removeTagFromTodos(todos []*model.Todo, tag string) {
	re := tagBoundaryRegex(tag)
	for _, t := range todos {
		if re.MatchString(t.Text) {
			t.Text = cleanupSpaces(re.ReplaceAllString(t.Text, ""))
			t.Category = t.ExtractCategory()
		}
	}
}

// renameTagInTodos renames #oldTag → #newTag (exact, word-boundary) in the text of all todos.
func renameTagInTodos(todos []*model.Todo, oldTag, newTag string) {
	re := tagBoundaryRegex(oldTag)
	newPattern := "#" + newTag
	for _, t := range todos {
		if re.MatchString(t.Text) {
			t.Text = re.ReplaceAllString(t.Text, newPattern)
			t.Category = t.ExtractCategory()
		}
	}
}

// cleanupSpaces collapses multiple consecutive spaces into one and trims.
func cleanupSpaces(s string) string {
	words := strings.Fields(s)
	return strings.Join(words, " ")
}

// filteredTodos returns only the todos matching activeFilter (or all if no filter).
func (m Model) filteredTodos() []*model.Todo {
	if m.activeFilter == "" {
		return m.todos
	}
	re := tagBoundaryRegex(m.activeFilter)
	var result []*model.Todo
	for _, t := range m.todos {
		if re.MatchString(t.Text) {
			result = append(result, t)
		}
	}
	return result
}

// renderFilterHeader returns the 2-line filter header string (empty if no filter).
func (m Model) renderFilterHeader() string {
	if m.activeFilter == "" {
		return ""
	}
	return activeFilterStyle.Render(fmt.Sprintf("  Filtered by: #%s", m.activeFilter)) + "\n\n"
}
