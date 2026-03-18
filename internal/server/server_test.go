package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/karimStekelenburg/dooing-tmux/internal/model"
)


func TestServerStartAndTodosEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a fixed test port. Pick something unlikely to conflict.
	port := 17291
	srv := New(port)

	todos := []*model.Todo{
		model.NewTodo("test task #work"),
	}
	srv.SetTodos(todos)

	errc, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	_ = errc

	// Give the listener a moment to be ready.
	time.Sleep(50 * time.Millisecond)

	url := fmt.Sprintf("http://127.0.0.1:%d/todos", port)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		t.Fatalf("GET /todos: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /todos status: got %d, want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type: got %q, want application/json", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	var loaded []*model.Todo
	if err := json.Unmarshal(body, &loaded); err != nil {
		t.Fatalf("parsing JSON: %v", err)
	}
	if len(loaded) != 1 {
		t.Errorf("got %d todos, want 1", len(loaded))
	}
	if loaded[0].Text != todos[0].Text {
		t.Errorf("todo text: got %q, want %q", loaded[0].Text, todos[0].Text)
	}
}

func TestServerRootEndpoint(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := 17292
	srv := New(port)
	srv.SetTodos([]*model.Todo{model.NewTodo("test")})

	_, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	url := fmt.Sprintf("http://127.0.0.1:%d/", port)
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		t.Fatalf("GET /: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET / status: got %d, want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type: got %q, want text/html", ct)
	}

	body, _ := io.ReadAll(resp.Body)
	bs := string(body)
	if !strings.Contains(bs, "dooing") {
		t.Error("root page should contain 'dooing'")
	}
	if !strings.Contains(bs, "/todos") {
		t.Error("root page should contain a link to /todos")
	}
}

func TestServerShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	port := 17293
	srv := New(port)
	srv.SetTodos([]*model.Todo{})

	errc, err := srv.Start(ctx)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Cancel context to trigger shutdown.
	cancel()

	select {
	case <-errc:
		// Server stopped — good.
	case <-time.After(2 * time.Second):
		t.Fatal("server did not shut down within 2s")
	}

	// Further requests should fail.
	url := fmt.Sprintf("http://127.0.0.1:%d/todos", port)
	_, err = http.Get(url) //nolint:noctx
	if err == nil {
		t.Error("expected error after server shutdown, got nil")
	}
}

func TestServerPortConflict(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port := 17294
	srv1 := New(port)
	_, err := srv1.Start(ctx)
	if err != nil {
		t.Fatalf("first Start: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	srv2 := New(port)
	_, err = srv2.Start(ctx)
	if err == nil {
		t.Error("expected error for port conflict, got nil")
	}
}

func TestLanIP(t *testing.T) {
	ip := lanIP()
	if ip == "" {
		t.Error("lanIP returned empty string")
	}
	// Should be either a valid IP or "localhost".
	if ip != "localhost" {
		parts := strings.Split(ip, ".")
		if len(parts) != 4 {
			t.Errorf("lanIP = %q: not a valid IPv4 or localhost", ip)
		}
	}
}

func TestSetTodos(t *testing.T) {
	srv := New(17295)
	todos := []*model.Todo{model.NewTodo("a"), model.NewTodo("b")}
	srv.SetTodos(todos)

	srv.mu.RLock()
	got := len(srv.todos)
	srv.mu.RUnlock()

	if got != 2 {
		t.Errorf("SetTodos: got %d todos, want 2", got)
	}
}
