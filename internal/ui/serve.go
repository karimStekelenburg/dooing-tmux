package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/karimStekelenburg/dooing-tmux/internal/server"
)

// serverErrMsg carries an error from the background HTTP server goroutine.
type serverErrMsg struct{ err error }

// toggleServer starts the HTTP share server if it is not running, or stops it
// if it is. Returns the updated model and any command to run.
func (m Model) toggleServer() (tea.Model, tea.Cmd) {
	if m.srv != nil {
		// Server already running — stop it.
		m.stopServer()
		m.statusMsg = "Share server stopped"
		return m, nil
	}

	port := m.cfg.ServePort
	if port <= 0 {
		port = 7283
	}

	srv := server.New(port)
	srv.SetTodos(m.todos)

	ctx, cancel := context.WithCancel(context.Background())
	errc, err := srv.Start(ctx)
	if err != nil {
		cancel()
		m.statusMsg = "Share server error: " + err.Error()
		return m, nil
	}

	m.srv = srv
	m.srvCancel = cancel
	m.statusMsg = "Sharing at http://" + srv.Addr() + "  (QR: http://" + srv.Addr() + "/)"

	// Forward any server error back as a tea.Msg so the user sees it.
	return m, func() tea.Msg {
		err := <-errc
		if err != nil {
			return serverErrMsg{err: err}
		}
		return serverErrMsg{}
	}
}

// stopServer shuts down the HTTP server if running.
func (m *Model) stopServer() {
	if m.srvCancel != nil {
		m.srvCancel()
		m.srvCancel = nil
	}
	m.srv = nil
}

// syncServerTodos pushes the current todos to the running server (if any).
func (m Model) syncServerTodos() {
	if m.srv != nil {
		m.srv.SetTodos(m.todos)
	}
}

// StartServer starts the HTTP share server and returns the updated model.
// This is intended for use from main.go when the --serve flag is provided.
func (m Model) StartServer() Model {
	newM, _ := m.toggleServer()
	return newM.(Model)
}
