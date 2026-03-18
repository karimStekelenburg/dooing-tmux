package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/karimStekelenburg/dooing-tmux/internal/ui"
)

func main() {
	project := flag.Bool("project", false, "use project-scoped todos (git root)")
	flag.Parse()

	m := ui.NewModel(*project)

	// When running inside a tmux popup (TMUX env var set), skip the alternate
	// screen so tmux provides the surrounding frame.
	opts := []tea.ProgramOption{tea.WithAltScreen()}
	if ui.InTmux() {
		opts = []tea.ProgramOption{}
	}

	p := tea.NewProgram(m, opts...)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
