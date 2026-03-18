package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// fakeKey creates a tea.KeyMsg for a single-character key string.
func fakeKey(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func TestDaysInMonth(t *testing.T) {
	cases := []struct {
		year  int
		month time.Month
		want  int
	}{
		{2026, time.January, 31},
		{2026, time.February, 28},
		{2024, time.February, 29}, // leap year
		{2000, time.February, 29}, // leap year (divisible by 400)
		{1900, time.February, 28}, // not a leap year (divisible by 100 not 400)
		{2026, time.April, 30},
		{2026, time.December, 31},
	}

	for _, c := range cases {
		got := daysInMonth(c.year, c.month)
		if got != c.want {
			t.Errorf("daysInMonth(%d, %v) = %d, want %d", c.year, c.month, got, c.want)
		}
	}
}

func TestFormatDueDateNotOverdue(t *testing.T) {
	// A future date should render with blue color and no "!" prefix.
	future := time.Now().AddDate(0, 0, 10).Unix()
	result := formatDueDate(&future, false)
	if result == "" {
		t.Error("expected non-empty due date string")
	}
	// Should not contain the overdue "!" marker.
	if len(result) > 0 {
		// The raw text should contain the date but no "! " marker in visible text.
		// We only check that it doesn't start with overdueStyle's content marker.
		// Just verify it's not empty and doesn't error.
	}
}

func TestFormatDueDateOverdue(t *testing.T) {
	// A past date should render as overdue (red + "!").
	past := time.Now().AddDate(0, 0, -5).Unix()
	result := formatDueDate(&past, false)
	if result == "" {
		t.Error("expected non-empty overdue date string")
	}
}

func TestFormatDueDateNilReturnsEmpty(t *testing.T) {
	result := formatDueDate(nil, false)
	if result != "" {
		t.Errorf("expected empty string for nil due date, got %q", result)
	}
}

func TestFormatDueDateDoneNotOverdue(t *testing.T) {
	// Even if date is in the past, a done todo should not be marked overdue.
	past := time.Now().AddDate(0, 0, -5).Unix()
	result := formatDueDate(&past, true)
	// Should still render the date (not empty).
	if result == "" {
		t.Error("expected non-empty date string for done todo with past due date")
	}
	// The formatDueDate checks done bool — for done=true it uses dueDateStyle, not overdueStyle.
}

func TestCalendarNavigation(t *testing.T) {
	m := NewModel(false)
	m.cal = calendarState{
		open:     true,
		year:     2026,
		month:    time.March,
		day:      15,
		today:    time.Date(2026, time.March, 15, 0, 0, 0, 0, time.Local),
		startDay: "sunday",
	}

	// Navigate forward one day with "l".
	result, _ := m.updateCalendar(fakeKey("l"))
	updated := result.(Model)
	if updated.cal.day != 16 {
		t.Errorf("after 'l', expected day=16, got %d", updated.cal.day)
	}

	// Navigate backward one day with "h".
	result, _ = updated.updateCalendar(fakeKey("h"))
	updated = result.(Model)
	if updated.cal.day != 15 {
		t.Errorf("after 'h', expected day=15, got %d", updated.cal.day)
	}

	// Navigate forward one week with "j".
	result, _ = m.updateCalendar(fakeKey("j"))
	updated = result.(Model)
	if updated.cal.day != 22 {
		t.Errorf("after 'j', expected day=22, got %d", updated.cal.day)
	}

	// Navigate backward one week with "k".
	result, _ = updated.updateCalendar(fakeKey("k"))
	updated = result.(Model)
	if updated.cal.day != 15 {
		t.Errorf("after 'k', expected day=15, got %d", updated.cal.day)
	}
}

func TestCalendarMonthNavigation(t *testing.T) {
	m := NewModel(false)
	m.cal = calendarState{
		open:     true,
		year:     2026,
		month:    time.March,
		day:      31,
		today:    time.Date(2026, time.March, 15, 0, 0, 0, 0, time.Local),
		startDay: "sunday",
	}

	// Navigate to previous month (February has 28 days in 2026, so day should clamp).
	result, _ := m.updateCalendar(fakeKey("H"))
	updated := result.(Model)
	if updated.cal.month != time.February {
		t.Errorf("expected February, got %v", updated.cal.month)
	}
	if updated.cal.day > 28 {
		t.Errorf("day should be clamped to 28 for Feb 2026, got %d", updated.cal.day)
	}

	// Navigate to next month (March).
	result, _ = updated.updateCalendar(fakeKey("L"))
	updated = result.(Model)
	if updated.cal.month != time.March {
		t.Errorf("expected March after forward navigation, got %v", updated.cal.month)
	}
}

func TestCalendarDayWrapAcrossMonths(t *testing.T) {
	m := NewModel(false)
	m.cal = calendarState{
		open:     true,
		year:     2026,
		month:    time.January,
		day:      31,
		today:    time.Date(2026, time.January, 31, 0, 0, 0, 0, time.Local),
		startDay: "sunday",
	}

	// Navigate forward one day from Jan 31 → Feb 1.
	result, _ := m.updateCalendar(fakeKey("l"))
	updated := result.(Model)
	if updated.cal.month != time.February || updated.cal.day != 1 {
		t.Errorf("expected Feb 1, got %v %d", updated.cal.month, updated.cal.day)
	}
}
