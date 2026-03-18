# dooing-tmux

A terminal-based todo list manager built with [Bubble Tea](https://github.com/charmbracelet/bubbletea), designed to work seamlessly inside tmux popups.

## Features

- Fast TUI todo management with keyboard-driven navigation
- Tag-based organization (`#tag` in todo text)
- Priority scoring and due dates
- Per-project todos (git-root scoped)
- Designed for `tmux display-popup -E`

## Installation

```bash
go install github.com/karimStekelenburg/dooing-tmux@latest
```

Or build from source:

```bash
git clone https://github.com/karimStekelenburg/dooing-tmux
cd dooing-tmux
make build
```

## Usage

```bash
# Global todos
dooing-tmux

# Project-scoped todos (uses git root)
dooing-tmux --project

# tmux popup (add to tmux.conf)
bind-key t display-popup -E -w 60 -h 22 "dooing-tmux"
```

## Keybindings

| Key | Action |
|-----|--------|
| `i` | Create new todo |
| `e` | Edit selected todo |
| `x` | Toggle status (pending / in-progress / done) |
| `d` | Delete todo |
| `u` | Undo last delete |
| `?` | Show help |
| `q` | Quit |

## Development

```bash
make build   # compile binary
make run     # run directly
make test    # run tests
make lint    # run golangci-lint
```

## Requirements

- Go 1.23+
- A terminal with true-color support (optional but recommended)
