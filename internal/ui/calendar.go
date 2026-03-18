package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// calendarState holds the state for the calendar popup overlay.
type calendarState struct {
	open     bool
	todoID   string // ID of the todo being edited
	year     int
	month    time.Month
	day      int       // currently highlighted day
	today    time.Time // cached today for highlighting
	startDay string    // "sunday" or "monday"
}

var (
	calendarBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("33")).
				Padding(0, 1)

	calendarTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("33"))

	calendarHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("243"))

	calendarTodayStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("33"))

	calendarSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Reverse(true)

	calendarFooterStyle = lipgloss.NewStyle().
				Faint(true)

	calendarWeekendStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("203"))

	overdueStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	dueDateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("33"))
)

// monthNames maps month number to abbreviated and full names (en).
var monthNames = [13]string{
	"", // 1-indexed
	"January", "February", "March", "April",
	"May", "June", "July", "August",
	"September", "October", "November", "December",
}

var monthAbbr = [13]string{
	"",
	"Jan", "Feb", "Mar", "Apr",
	"May", "Jun", "Jul", "Aug",
	"Sep", "Oct", "Nov", "Dec",
}

// weekdayHeaderSunday is the header row when week starts on Sunday.
var weekdayHeaderSunday = "Su Mo Tu We Th Fr Sa"

// weekdayHeaderMonday is the header row when week starts on Monday.
var weekdayHeaderMonday = "Mo Tu We Th Fr Sa Su"

// openCalendar opens the calendar for the todo at the current cursor.
func (m *Model) openCalendar() {
	visible := m.filteredTodos()
	if len(visible) == 0 {
		return
	}
	t := visible[m.cursor]
	now := time.Now()
	year, month, day := now.Year(), now.Month(), now.Day()

	// If todo has a due date, start there.
	if t.DueAt != nil {
		d := time.Unix(*t.DueAt, 0)
		year, month, day = d.Year(), d.Month(), d.Day()
	}

	m.cal = calendarState{
		open:     true,
		todoID:   t.ID,
		year:     year,
		month:    month,
		day:      day,
		today:    time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		startDay: m.cfg.Calendar.StartDay,
	}
}

// updateCalendar handles input when the calendar popup is open.
func (m Model) updateCalendar(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "q", "esc":
		m.cal.open = false

	case "h":
		// Previous day.
		d := time.Date(m.cal.year, m.cal.month, m.cal.day-1, 0, 0, 0, 0, time.Local)
		m.cal.year, m.cal.month, m.cal.day = d.Year(), d.Month(), d.Day()

	case "l":
		// Next day.
		d := time.Date(m.cal.year, m.cal.month, m.cal.day+1, 0, 0, 0, 0, time.Local)
		m.cal.year, m.cal.month, m.cal.day = d.Year(), d.Month(), d.Day()

	case "k":
		// Previous week.
		d := time.Date(m.cal.year, m.cal.month, m.cal.day-7, 0, 0, 0, 0, time.Local)
		m.cal.year, m.cal.month, m.cal.day = d.Year(), d.Month(), d.Day()

	case "j":
		// Next week.
		d := time.Date(m.cal.year, m.cal.month, m.cal.day+7, 0, 0, 0, 0, time.Local)
		m.cal.year, m.cal.month, m.cal.day = d.Year(), d.Month(), d.Day()

	case "H":
		// Previous month.
		d := time.Date(m.cal.year, m.cal.month-1, 1, 0, 0, 0, 0, time.Local)
		m.cal.year, m.cal.month = d.Year(), d.Month()
		// Clamp day to valid range.
		maxDay := daysInMonth(m.cal.year, m.cal.month)
		if m.cal.day > maxDay {
			m.cal.day = maxDay
		}

	case "L":
		// Next month.
		d := time.Date(m.cal.year, m.cal.month+1, 1, 0, 0, 0, 0, time.Local)
		m.cal.year, m.cal.month = d.Year(), d.Month()
		maxDay := daysInMonth(m.cal.year, m.cal.month)
		if m.cal.day > maxDay {
			m.cal.day = maxDay
		}

	case "enter":
		// Select the current day — set due date to 23:59:59 of that day.
		selected := time.Date(m.cal.year, m.cal.month, m.cal.day, 23, 59, 59, 0, time.Local)
		ts := selected.Unix()
		for _, t := range m.todos {
			if t.ID == m.cal.todoID {
				t.DueAt = &ts
				break
			}
		}
		m.sortTodos()
		_ = m.st.Save(m.storePath, m.todos)
		m.cal.open = false
		m.statusMsg = fmt.Sprintf("Due date set to %s %d, %d",
			monthAbbr[m.cal.month], m.cal.day, m.cal.year)
	}

	return m, nil
}

// removeDueDate clears the due date from the selected todo.
func (m *Model) removeDueDate() {
	visible := m.filteredTodos()
	if len(visible) == 0 {
		return
	}
	t := visible[m.cursor]
	for _, todo := range m.todos {
		if todo.ID == t.ID {
			todo.DueAt = nil
			break
		}
	}
	m.sortTodos()
	_ = m.st.Save(m.storePath, m.todos)
	m.statusMsg = "Due date removed"
}

// renderCalendar returns the styled calendar overlay string.
func (m Model) renderCalendar() string {
	cal := m.cal
	year, month, day := cal.year, cal.month, cal.day

	var sb strings.Builder
	// Title: month name + year.
	title := fmt.Sprintf(" %s %d ", monthNames[month], year)
	sb.WriteString(calendarTitleStyle.Render(title))
	sb.WriteString("\n\n")

	// Weekday header.
	startMonday := strings.ToLower(cal.startDay) == "monday"
	if startMonday {
		sb.WriteString(calendarHeaderStyle.Render(weekdayHeaderMonday))
	} else {
		sb.WriteString(calendarHeaderStyle.Render(weekdayHeaderSunday))
	}
	sb.WriteString("\n")

	// First weekday of the month.
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.Local).Weekday()
	// offset: how many blank cells before day 1.
	var offset int
	if startMonday {
		offset = (int(first) + 6) % 7 // Mon=0..Sun=6
	} else {
		offset = int(first) // Sun=0..Sat=6
	}

	// Build day grid.
	daysCount := daysInMonth(year, month)
	col := 0

	// Print leading blanks.
	for i := 0; i < offset; i++ {
		sb.WriteString("   ")
		col++
	}

	for d := 1; d <= daysCount; d++ {
		cell := fmt.Sprintf("%2d ", d)
		weekday := (offset + d - 1) % 7

		var isWeekend bool
		if startMonday {
			isWeekend = weekday == 5 || weekday == 6 // Sa/Su
		} else {
			isWeekend = weekday == 0 || weekday == 6 // Su/Sa
		}

		isToday := year == cal.today.Year() && month == cal.today.Month() && d == cal.today.Day()
		isSelected := d == day

		var rendered string
		switch {
		case isSelected:
			rendered = calendarSelectedStyle.Render(cell)
		case isToday:
			rendered = calendarTodayStyle.Render(cell)
		case isWeekend:
			rendered = calendarWeekendStyle.Render(cell)
		default:
			rendered = cell
		}

		sb.WriteString(rendered)
		col++
		if col == 7 {
			sb.WriteString("\n")
			col = 0
		}
	}
	if col != 0 {
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(calendarFooterStyle.Render("h/l day  j/k week  H/L month  enter select  q cancel"))

	return calendarBorderStyle.Render(sb.String())
}

// daysInMonth returns the number of days in the given month/year.
func daysInMonth(year int, month time.Month) int {
	// Day 0 of next month = last day of current month.
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// formatDueDate returns the formatted due date string for a todo.
// Returns "" if no due date.
func formatDueDate(dueAt *int64, done bool) string {
	if dueAt == nil {
		return ""
	}
	d := time.Unix(*dueAt, 0)
	label := fmt.Sprintf("[%s %d, %d]", monthAbbr[d.Month()], d.Day(), d.Year())

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()

	if !done && *dueAt < startOfToday {
		return overdueStyle.Render("[! " + label[1:len(label)-1] + "]")
	}
	return dueDateStyle.Render(label)
}
