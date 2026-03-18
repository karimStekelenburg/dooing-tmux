package ui

import (
	"testing"
)

func TestParseTimeEstimation(t *testing.T) {
	cases := []struct {
		input  string
		wantH  float64
		wantOK bool
	}{
		{"30m", 0.5, true},
		{"60m", 1.0, true},
		{"2h", 2.0, true},
		{"0.5h", 0.5, true},
		{"1d", 8.0, true},
		{"0.5d", 4.0, true},
		{"1w", 40.0, true},
		{"0.5w", 20.0, true},
		{"", 0, false},
		{"abc", 0, false},
		{"5x", 0, false},
		{"-1h", 0, false},
		{"0h", 0, false},
	}

	for _, c := range cases {
		got, ok := parseTimeEstimation(c.input)
		if ok != c.wantOK {
			t.Errorf("parseTimeEstimation(%q): ok=%v, want %v", c.input, ok, c.wantOK)
			continue
		}
		if ok && abs(got-c.wantH) > 0.0001 {
			t.Errorf("parseTimeEstimation(%q): hours=%.4f, want %.4f", c.input, got, c.wantH)
		}
	}
}

func TestFormatTimeEstimation(t *testing.T) {
	cases := []struct {
		hours float64
		want  string
	}{
		{0.5, "[≈ 30m]"},   // 30 minutes (< 1h)
		{1.0, "[≈ 1h]"},    // 1 hour (< 8h)
		{2.0, "[≈ 2h]"},    // 2 hours (< 8h)
		{4.0, "[≈ 4h]"},    // 4 hours (< 8h)
		{8.0, "[≈ 1d]"},    // 1 day (< 40h)
		{16.0, "[≈ 2d]"},   // 2 days (< 40h)
		{40.0, "[≈ 1w]"},   // 1 week (>= 40h)
		{80.0, "[≈ 2w]"},   // 2 weeks
		{0, ""},            // zero = empty
	}

	for _, c := range cases {
		got := formatTimeEstimation(c.hours)
		if got != c.want {
			t.Errorf("formatTimeEstimation(%.1f) = %q, want %q", c.hours, got, c.want)
		}
	}
}

// TestRoundtrip verifies that parse→format gives back the correct unit display.
func TestTimeEstimationRoundtrip(t *testing.T) {
	cases := []struct {
		input       string
		wantDisplay string
	}{
		{"30m", "[≈ 30m]"},
		{"120m", "[≈ 2h]"},   // 120m = 2h (< 8h)
		{"2h", "[≈ 2h]"},
		{"1d", "[≈ 1d]"},     // 1d = 8h (< 40h)
		{"1w", "[≈ 1w]"},     // 1w = 40h (>= 40h)
		{"5w", "[≈ 5w]"},     // 5w = 200h (>= 40h)
	}

	for _, c := range cases {
		hours, ok := parseTimeEstimation(c.input)
		if !ok {
			t.Errorf("parse(%q) failed", c.input)
			continue
		}
		got := formatTimeEstimation(hours)
		if got != c.wantDisplay {
			t.Errorf("roundtrip(%q): got %q, want %q (hours=%.4f)", c.input, got, c.wantDisplay, hours)
		}
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
