// Package server provides a read-only HTTP server for sharing todos over LAN
// via a QR code.
package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	qrcode "github.com/skip2/go-qrcode"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)

// Server exposes the current todo list over HTTP on a configurable port.
type Server struct {
	port   int
	mu     sync.RWMutex
	todos  []*model.Todo
	httpSrv *http.Server
	addr   string // resolved "host:port" after Start
}

// New creates a new Server that will listen on the given port.
func New(port int) *Server {
	s := &Server{port: port}
	mux := http.NewServeMux()
	mux.HandleFunc("/todos", s.handleTodos)
	mux.HandleFunc("/", s.handleRoot)
	s.httpSrv = &http.Server{Handler: mux}
	return s
}

// SetTodos atomically replaces the todo list served by the HTTP endpoints.
func (s *Server) SetTodos(todos []*model.Todo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]*model.Todo, len(todos))
	copy(cp, todos)
	s.todos = cp
}

// Addr returns the resolved "host:port" string, available after Start.
func (s *Server) Addr() string {
	return s.addr
}

// Start begins listening in a background goroutine. The ctx is used for graceful
// shutdown — when it is cancelled the server stops accepting new connections.
// The returned error channel receives at most one value (nil on clean shutdown, non-nil otherwise).
func (s *Server) Start(ctx context.Context) (<-chan error, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return nil, fmt.Errorf("listen on port %d: %w", s.port, err)
	}

	lanIP := lanIP()
	s.addr = fmt.Sprintf("%s:%d", lanIP, s.port)

	errc := make(chan error, 1)
	go func() {
		errc <- s.httpSrv.Serve(ln)
	}()

	// Shutdown when ctx is cancelled.
	go func() {
		<-ctx.Done()
		_ = s.httpSrv.Shutdown(context.Background())
	}()

	return errc, nil
}

func (s *Server) handleTodos(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	todos := s.todos
	s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if todos == nil {
		todos = []*model.Todo{}
	}
	_ = enc.Encode(todos)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	todosURL := fmt.Sprintf("http://%s/todos", s.addr)

	// Generate QR code as PNG, then embed as inline SVG-like data URI.
	var qrPNG []byte
	var qrErr string
	pngBytes, err := qrcode.Encode(todosURL, qrcode.Medium, 256)
	if err != nil {
		qrErr = err.Error()
	} else {
		qrPNG = pngBytes
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var buf bytes.Buffer
	buf.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>dooing – share todos</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 480px; margin: 3rem auto; padding: 0 1rem; text-align: center; background: #0f0f0f; color: #e0e0e0; }
  h1 { color: #9b8fef; margin-bottom: 0.25rem; }
  p  { color: #999; margin-top: 0; }
  .url { font-size: 1rem; background: #1e1e1e; border-radius: 6px; padding: 0.6rem 1rem; display: inline-block; word-break: break-all; color: #86e0b0; margin: 1rem 0; }
  img { border: 4px solid #fff; border-radius: 8px; margin-top: 1rem; }
  .err { color: #f87171; }
</style>
</head>
<body>
<h1>dooing</h1>
<p>Scan to read your todos on another device</p>
<div class="url">`)
	buf.WriteString(todosURL)
	buf.WriteString(`</div>`)
	if qrErr != "" {
		buf.WriteString(`<p class="err">QR generation failed: ` + qrErr + `</p>`)
	} else {
		// Embed PNG as base64 data URI.
		buf.WriteString(`<br><img src="data:image/png;base64,`)
		buf.WriteString(base64.StdEncoding.EncodeToString(qrPNG))
		buf.WriteString(`" alt="QR code" width="256" height="256">`)
	}
	buf.WriteString(`
</body>
</html>`)
	_, _ = w.Write(buf.Bytes())
}

// lanIP returns the first non-loopback IPv4 address of the host.
// Falls back to "localhost" if none is found.
func lanIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "localhost"
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ip4 := ip.To4(); ip4 != nil {
				return ip4.String()
			}
		}
	}
	return "localhost"
}

