package ui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// gitRoot runs `git rev-parse --show-toplevel` and returns the absolute path
// to the git repository root. Returns ("", error) when not in a git repo.
func gitRoot() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return strings.TrimSpace(string(out)), nil
}

// projectStorePath resolves the path for the project-scoped todos file.
// Returns ("", error) if not in a git repo.
func projectStorePath(filename string) (string, error) {
	root, err := gitRoot()
	if err != nil {
		return "", err
	}
	if filename == "" {
		filename = "dooing.json"
	}
	return filepath.Join(root, filename), nil
}

// ensureGitignore appends filename to .gitignore in root if not already present.
// mode: "true" → always, "false" → never, "prompt" → ask (falls back to no-op in non-interactive).
func ensureGitignore(root, filename, mode string) {
	switch strings.ToLower(mode) {
	case "false":
		return
	case "true":
		appendToGitignore(root, filename)
	default:
		// "prompt" or anything else: silently skip (no interactive prompt in TUI startup).
		// The user can manually add to .gitignore.
	}
}

// appendToGitignore adds filename to <root>/.gitignore if not already listed.
func appendToGitignore(root, filename string) {
	giPath := filepath.Join(root, ".gitignore")

	// Check if already present.
	if f, err := os.Open(giPath); err == nil {
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			if strings.TrimSpace(scanner.Text()) == filename {
				_ = f.Close()
				return // already present
			}
		}
		_ = f.Close()
	}

	// Append.
	f, err := os.OpenFile(giPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close() //nolint:errcheck
	_, _ = fmt.Fprintf(f, "\n# dooing-tmux project todos\n%s\n", filename)
}

// InTmux returns true when the process is running inside a tmux session.
// When true, the caller may want to skip the alternate screen to let tmux
// provide the frame (e.g. display-popup -E).
func InTmux() bool {
	return os.Getenv("TMUX") != ""
}
