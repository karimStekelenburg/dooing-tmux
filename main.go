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
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
