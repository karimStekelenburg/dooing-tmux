package ui

import (
	"testing"

	"github.com/karimStekelenburg/dooing-tmux/internal/config"
)

func TestResolvePriorityColor(t *testing.T) {
	groups := map[string]config.PriorityGroup{
		"high":   {Members: []string{"important", "urgent"}, Color: "#ff0000"},
		"medium": {Members: []string{"important"}, Color: "#ffff00"},
		"low":    {Members: []string{"urgent"}, Color: "#0000ff"},
	}

	tests := []struct {
		name       string
		priorities []string
		want       string
	}{
		{"both priorities matches high", []string{"important", "urgent"}, "#ff0000"},
		{"only important matches medium", []string{"important"}, "#ffff00"},
		{"only urgent matches low", []string{"urgent"}, "#0000ff"},
		{"no priorities returns empty", nil, ""},
		{"unknown priority returns empty", []string{"trivial"}, ""},
		{"empty groups", []string{"important"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := groups
			if tt.name == "empty groups" {
				g = nil
			}
			got := resolvePriorityColor(tt.priorities, g)
			if got != tt.want {
				t.Errorf("resolvePriorityColor(%v) = %q, want %q", tt.priorities, got, tt.want)
			}
		})
	}
}
