package ui

import (
	"math"
	"testing"
)

func TestParseTimeInput(t *testing.T) {
	tests := []struct {
		input   string
		wantH   float64
		wantErr bool
	}{
		{"30m", 0.5, false},
		{"60m", 1.0, false},
		{"2h", 2.0, false},
		{"1d", 8.0, false},
		{"0.5w", 20.0, false},
		{"1w", 40.0, false},
		{"", 0, true},         // too short
		{"x", 0, true},        // too short
		{"2x", 0, true},       // unknown unit
		{"-1h", 0, true},      // non-positive
		{"abch", 0, true},     // non-numeric
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseTimeInput(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseTimeInput(%q) = %f, want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("parseTimeInput(%q) unexpected error: %v", tt.input, err)
				return
			}
			if math.Abs(got-tt.wantH) > 0.001 {
				t.Errorf("parseTimeInput(%q) = %f, want %f", tt.input, got, tt.wantH)
			}
		})
	}
}

func TestFormatHours(t *testing.T) {
	tests := []struct {
		hours float64
		want  string
	}{
		{0.5, "30m"},
		{1.0, "1h"},
		{2.0, "2h"},
		{8.0, "1d"},
		{4.0, "4h"},   // < 8h → hours
		{16.0, "2d"},
		{40.0, "1w"},
		{80.0, "2w"},
		{20.0, "2.5d"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatHours(tt.hours)
			if got != tt.want {
				t.Errorf("formatHours(%f) = %q, want %q", tt.hours, got, tt.want)
			}
		})
	}
}

func TestRenderTimeEstimate(t *testing.T) {
	if renderTimeEstimate(0) != "" {
		t.Error("expected empty string for zero hours")
	}
	got := renderTimeEstimate(0.5)
	if got == "" {
		t.Error("expected non-empty string for 0.5 hours")
	}
}
