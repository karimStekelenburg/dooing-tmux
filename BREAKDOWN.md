# Dooing-Tmux: Complete Feature Breakdown

> Reverse-engineered from [atiladefreitas/dooing](https://github.com/atiladefreitas/dooing) — a minimalist Neovim todo manager.
> Goal: Reimplement as a standalone TUI app launchable via `tmux popup`.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [Data Model](#2-data-model)
3. [Persistence Layer](#3-persistence-layer)
4. [Core Operations (CRUD + State)](#4-core-operations)
5. [Tag System](#5-tag-system)
6. [Priority System](#6-priority-system)
7. [Sorting Algorithm](#7-sorting-algorithm)
8. [Nested Tasks](#8-nested-tasks)
9. [Due Dates & Calendar](#9-due-dates--calendar)
10. [Time Estimation](#10-time-estimation)
11. [Search & Filtering](#11-search--filtering)
12. [Undo System](#12-undo-system)
13. [Import / Export](#13-import--export)
14. [Scratchpad / Notes](#14-scratchpad--notes)
15. [UI Layout & Rendering](#15-ui-layout--rendering)
16. [Keybindings](#16-keybindings)
17. [Views & Sub-Windows](#17-views--sub-windows)
18. [Highlights & Theming](#18-highlights--theming)
19. [Per-Project Support](#19-per-project-support)
20. [Notifications](#20-notifications)
21. [QR Sharing (HTTP Server)](#21-qr-sharing)
22. [Language Recommendations](#22-language-recommendations)

---

## 1. Architecture Overview

```
┌─────────────────────────────────────────────┐
│                  CLI Entry                   │
│         (tmux popup / direct run)            │
├─────────────────────────────────────────────┤
│              Config Layer                    │
│  (defaults + user overrides, TOML or JSON)  │
├──────────────┬──────────────────────────────┤
│  State/Data  │         UI Layer             │
│  - CRUD      │  - Main window (todo list)   │
│  - Sorting   │  - Tag picker                │
│  - Tags      │  - Search results            │
│  - Undo      │  - Calendar picker           │
│  - Persist   │  - Help overlay              │
│  - Search    │  - Priority selector         │
│              │  - Confirmation dialog        │
│              │  - Scratchpad editor          │
│              │  - Due notifications          │
│              │  - Quick keys panel           │
└──────────────┴──────────────────────────────┘
```

The original has clean separation: `state.lua` (data), `config.lua` (options), and `ui/` (rendering + interaction). Mirror this in your implementation.

---

## 2. Data Model

Each todo is a flat object. The entire collection is a JSON array.

```json
{
  "id": "1679000000_1234",
  "text": "Fix login bug #backend",
  "done": false,
  "in_progress": false,
  "category": "backend",
  "created_at": 1679000000,
  "completed_at": null,
  "priorities": ["urgent", "important"],
  "estimated_hours": 2.0,
  "due_at": 1706140799,
  "notes": "Check the OAuth flow\nMight be a token expiry issue",
  "parent_id": null,
  "depth": 0
}
```

### Field Reference

| Field | Type | Description |
|---|---|---|
| `id` | string | `"{unix_timestamp}_{random_1000_9999}"` — unique identifier |
| `text` | string | Todo text, may contain inline `#tags` |
| `done` | bool | Completion state |
| `in_progress` | bool | In-progress state (tri-state with `done`) |
| `category` | string\|null | First `#tag` found in text at creation time |
| `created_at` | int | Unix timestamp of creation |
| `completed_at` | int\|null | Unix timestamp when marked done |
| `priorities` | string[]\|null | List of assigned priority names |
| `estimated_hours` | float\|null | Estimated time in hours |
| `due_at` | int\|null | Unix timestamp (set to 23:59:59 of chosen day) |
| `notes` | string | Scratchpad/notes content (markdown) |
| `parent_id` | string\|null | ID of parent todo (for nesting) |
| `depth` | int | Nesting level (0 = top-level) |

### State Machine: Todo Status

```
    ┌──────────┐     toggle      ┌─────────────┐     toggle      ┌────────┐
    │ Pending  │ ──────────────> │ In Progress │ ──────────────> │  Done  │
    │ done=F   │                 │ done=F      │                 │ done=T │
    │ prog=F   │                 │ prog=T      │                 │ prog=F │
    └──────────┘                 └─────────────┘                 └────────┘
         ^                                                            │
         └────────────────────── toggle ──────────────────────────────┘
                              (clears completed_at)
```

On transition to Done: set `completed_at = now()`. On transition to Pending: clear `completed_at`.

---

## 3. Persistence Layer

### File Locations
- **Global:** `$XDG_DATA_HOME/dooing/todos.json` (original uses Neovim's `stdpath("data")`)
- **Per-project:** `<git_root>/dooing.json` (configurable filename)

### Format
Plain JSON array. The entire array is serialized/deserialized atomically (no incremental saves).

### Pretty Printing (optional)
Shell out to `jq .` or `python3 -m json.tool` for human-readable files.

### Migration
On load, add missing fields to old todos (backfill `id`, `parent_id`, `depth`).

---

## 4. Core Operations

### Create
```
add_todo(text, priority_names?) → Todo
  - Generate id = "{time}_{random}"
  - Extract category = first #tag from text
  - Append to array
  - Save to disk
```

### Read
Direct array access. All filtering/sorting happens at render time.

### Update (Toggle)
```
toggle_todo(index) →
  if pending      → set in_progress=true
  if in_progress  → set done=true, in_progress=false, completed_at=now()
  if done         → set done=false, completed_at=nil
  Save to disk
```

### Update (Edit)
Prompt user for new text, replace `todo.text`, re-extract category. Save.

### Delete
```
delete_todo(index) →
  if todo.done → delete immediately
  if !todo.done → show confirmation dialog, delete on "Y"
  Store in undo history before removal
  Save to disk
```

### Delete Completed
Remove all `done=true` todos. Each stored in undo history. For nested tasks: orphaned children of deleted parents are promoted to depth=0, parent_id=nil.

### Deduplication
Hash each todo via SHA-256 of its serialized form. Remove duplicates (keep first occurrence).

---

## 5. Tag System

Tags are **inline in the text**, prefixed with `#`. NOT stored as a separate list.

- **Extraction:** regex `#(\w+)` matches tags
- **Category field:** set to first match at creation time only (not updated on edit)
- **Tag listing:** scan all todos for unique `#tags`, return sorted
- **Rename tag:** find-and-replace `#oldname` → `#newname` across all todos
- **Delete tag:** remove `#tagname` text from all todos (does not delete the todos)
- **Filter by tag:** set `active_filter = "tagname"`, rendering skips non-matching todos

---

## 6. Priority System

### Configuration

```json
{
  "priorities": [
    { "name": "important", "weight": 4 },
    { "name": "urgent", "weight": 2 }
  ],
  "priority_groups": {
    "high":   { "members": ["important", "urgent"], "color": "red" },
    "medium": { "members": ["important"],           "color": "yellow" },
    "low":    { "members": ["urgent"],              "color": "blue" }
  },
  "hour_score_value": 0.125
}
```

### Assignment
Each todo can have zero or more priorities from the configured list. Assigned via a multi-select checkbox UI.

### Scoring Algorithm

```
score(todo):
  if todo.done → return 0
  base = sum(weight for each priority in todo.priorities)
  if todo.estimated_hours:
    multiplier = 1 / (estimated_hours * hour_score_value)
  else:
    multiplier = 1
  return base * multiplier
```

This creates a "quick wins first" bias — shorter tasks with high priority score highest.

### Group Resolution (for coloring)
Groups are checked largest-first (most members). The first group whose `members` are ALL present in the todo's priority list wins. Example: a todo with `["important", "urgent"]` matches `high` before `medium`.

---

## 7. Sorting Algorithm

Todos are sorted on every render. Multi-key sort:

1. **Completion status:** undone before done
2. **Completed time** (optional config): most recently completed first among done todos
3. **Priority score:** higher score first (descending)
4. **Due date:** earlier due date first; todos WITH due dates before those WITHOUT
5. **Creation time:** earlier first

### Structure-Aware Sort (nested tasks)
When nesting is enabled:
1. Extract top-level todos (depth=0), sort them by the rules above
2. For each top-level todo, gather its children recursively (depth-first)
3. Sort children within each parent group independently
4. Reconstruct flat array preserving depth-first tree order

---

## 8. Nested Tasks

- **Creation:** `add_nested_todo(text, parent_index)` — sets `parent_id` to parent's `id`, `depth = parent.depth + 1`
- **Insertion position:** immediately after parent and its existing children
- **Rendering:** indented by `depth * indent_size` spaces (default `indent_size = 2`)
- **Folding:** fold state tracked by todo `id`, preserved across re-renders
- **Deletion of parent:** orphaned children promoted to top-level (`depth=0, parent_id=nil`)
- **Sort preservation:** parent-child hierarchy maintained during sort (children sorted within their group, not globally)
- **Completion cascade:** optional `move_completed_to_end` moves done items to bottom within their nesting group

---

## 9. Due Dates & Calendar

### Calendar Widget
A small popup (26×9 characters) showing a month grid:

```
┌──── January 2026 ────┐
│                       │
│ Su Mo Tu We Th Fr Sa  │
│           1  2  3  4  │
│  5  6  7  8  9 10 11  │
│ 12 13 14 15 16 17 18  │
│ 19 20 21 22 23 24 25  │
│ 26 27 28 29 30 31     │
│                       │
└───────────────────────┘
```

- **Navigation:** h/l (day), j/k (week), H/L (month)
- **Selection:** Enter confirms, q cancels
- **7 languages:** en, pt, es, fr, de, it, jp (month names + weekday abbreviations)
- **Configurable start day:** Sunday or Monday
- **Today highlighting:** distinct color for today's date vs selected date vs regular
- **Date storage:** `due_at` = Unix timestamp at 23:59:59 of selected day

### Due Date Display
- Normal: `[📅 January 15, 2026]` (calendar icon configurable)
- Overdue: `[!📅 January 15, 2026]` with error/red highlight
- Overdue = `due_at < start_of_today` (where start_of_today = midnight of current day)

---

## 10. Time Estimation

### Input Format
User enters strings like: `15m`, `2h`, `1d`, `0.5w`

### Parsing
```
parse("15m")  → 0.25 hours
parse("2h")   → 2.0 hours
parse("1d")   → 8.0 hours    (1 day = 8 working hours)
parse("0.5w") → 20.0 hours   (1 week = 40 working hours)
```

### Display
Best-fit unit: `[≈ 15m]`, `[≈ 2h]`, `[≈ 1d]`, `[≈ 0.5w]`

### Impact
Factors into priority scoring via `hour_score_value` multiplier (see Priority System).

---

## 11. Search & Filtering

### Tag Filter
- Set `active_filter = "tagname"` → rendering skips todos not containing `#tagname`
- Adds 2-line header to buffer: `"  Filtered by: #tagname"` + blank line
- Clear with `c` key

### Text Search
- Case-insensitive substring match on `todo.text`
- Returns list of `{line_number, todo}` pairs
- Results shown in a separate popup window
- Enter on a result jumps to that todo in the main window

---

## 12. Undo System

- **In-memory stack** (not persisted across sessions)
- **Max history:** 100 items
- **Stored on every delete:** `{todo: deep_copy, original_index, timestamp}`
- **Undo:** pops most recent, re-inserts at original index (clamped to array bounds)
- **Scope:** only deletions are undoable (not edits, toggles, or other mutations)

---

## 13. Import / Export

- **Import:** reads JSON file → decodes → appends all todos to current list → sort → save
- **Export:** encodes current todos → writes to JSON file
- Both prompt for file path with filesystem completion
- Import MERGES (appends), does not replace

---

## 14. Scratchpad / Notes

- Per-todo markdown notes stored in `todo.notes`
- Opens in a floating window (60% of screen width and height)
- Configurable syntax highlighting (default: markdown)
- Auto-saves on window close/leave
- Enter and Escape both save and close
- Notes icon `󱞁` (nerd font) shown in todo line when notes exist

---

## 15. UI Layout & Rendering

### Main Window

- **Default size:** 55 wide × 20 tall
- **Border:** rounded (configurable: rounded, single, double, shadow, none)
- **Position:** center (configurable: center, right, left, top, bottom, top-right, top-left, bottom-right, bottom-left)
- **Title:** `" Global to-dos "` or `" <project> to-dos "` (centered)
- **Footer:** `" [?] for help "` (centered)
- **Non-editable:** buffer is set to non-modifiable; all interaction via keymaps

### Line Format

Each todo line is constructed from a configurable format array:

```
Default format: ["notes_icon", "icon", "text", "ect", "due_date", "relative_time"]
```

Rendered as:

```
  [indent][notes_icon] [icon] [text] [ect] [due_date]          [relative_time]
  ^^                                                             right-aligned
  2 spaces base indent
```

### Icons

| State | Icon | Unicode |
|---|---|---|
| Pending | ○ | U+25CB |
| In Progress | ◐ | U+25D0 |
| Done | ✓ | U+2713 |
| Notes present | 󱞁 | Nerd Font icon |

### Concrete Rendering Examples

```
  ○ Fix login bug #backend [≈ 2h] [January 20, 2026]             @3h ago
  ◐ Review PR #frontend                                          @1d ago
  ✓ Update docs #docs                                            @2d ago
    ○ Write API section #docs [≈ 30m]                             @5m ago
    ✓ Fix typos #docs                                            @1h ago
```

### Relative Timestamps
Right-aligned. Format: `@<N><unit> ago` or `just now`
- Units: `s` (seconds, <60s), `m` (minutes), `h` (hours), `d` (days), `w` (weeks), `mo` (months)

### Quick Keys Panel
Optional small window below the main window (not focusable). Two-column layout:

```
┌─────────────── Quick Keys ───────────────┐
│ i - New todo         T - Add time        │
│ <leader>tn - Nested  t - Tags            │
│ x - Toggle           / - Search          │
│ d - Delete           I - Import          │
│ u - Undo delete      E - Export          │
│ H - Add due date                         │
└──────────────────────────────────────────┘
```

---

## 16. Keybindings

### Main Window

| Key | Action |
|---|---|
| `i` | New todo |
| `n` | New nested todo (sub-task under cursor) |
| `x` | Toggle todo state (pending → in_progress → done → pending) |
| `d` | Delete todo (with confirmation for incomplete) |
| `D` | Delete all completed todos |
| `u` | Undo last delete |
| `e` | Edit todo text |
| `p` | Edit priorities (multi-select) |
| `H` | Add/change due date (opens calendar) |
| `r` | Remove due date |
| `T` | Add/change time estimation |
| `R` | Remove time estimation |
| `?` | Toggle help window |
| `t` | Toggle tags window |
| `c` | Clear tag filter |
| `/` | Search todos |
| `I` | Import from file |
| `E` | Export to file |
| `s` | Open scratchpad for todo under cursor |
| `f` | Reload from disk |
| `q` | Close window |

### Tags Window

| Key | Action |
|---|---|
| `Enter` | Filter by selected tag |
| `e` | Rename tag |
| `d` | Delete tag from all todos |
| `q` | Close |

### Calendar

| Key | Action |
|---|---|
| `h` / `l` | Previous / next day |
| `j` / `k` | Next / previous week |
| `H` / `L` | Previous / next month |
| `Enter` | Select date |
| `q` | Close |

### Search Results

| Key | Action |
|---|---|
| `Enter` | Jump to todo |
| `q` | Close |

### Priority Selector

| Key | Action |
|---|---|
| `Space` | Toggle checkbox |
| `Enter` | Confirm selection |
| `q` | Cancel |

### Confirmation Dialog

| Key | Action |
|---|---|
| `y` / `Y` | Confirm delete |
| `n` / `N` / `q` / `Esc` | Cancel |

---

## 17. Views & Sub-Windows

| Window | Size | Position | Purpose |
|---|---|---|---|
| Main | 55×20 | configurable (9 positions) | Todo list |
| Quick Keys | same width × ~8 | below main | Key reference |
| Help | 50×45 | right of center, z=100 | Full keymap reference |
| Tags | 30×10 | left of main | Tag list + filter |
| Search Results | 40×10 | left of main | Search matches |
| Calendar | 26×9 | relative to cursor | Date picker |
| Priority Selector | 40×auto | center | Multi-select checkboxes |
| Confirmation | 60×auto | center | Y/N delete confirmation |
| Scratchpad | 60%×60% | center | Per-todo notes editor |
| Due Notifications | 55×auto | center, z=50 | Overdue + due-today list |

---

## 18. Highlights & Theming

The original uses Neovim highlight groups linked to semantic groups so colors adapt to any colorscheme.

### Color Mapping for TUI

| Semantic | Original HL Group | Suggested TUI Color |
|---|---|---|
| Pending todo | `DooingPending` → `Question` | Cyan / Blue |
| In-progress todo | (same as Pending) | Yellow |
| Done todo | `DooingDone` → `Comment` | Dim / Gray |
| Tags | `Type` | Green |
| Filter header | `WarningMsg` | Yellow / Bold |
| Overdue due date | `ErrorMsg` | Red / Bold |
| Timestamps | `DooingTimestamp` → `Comment` | Dim / Gray |
| Priority High | `DiagnosticError` | Red |
| Priority Medium | `DiagnosticWarn` | Yellow |
| Priority Low | `DiagnosticInfo` | Blue |
| Calendar today | `Directory` | Blue |
| Calendar selected | `Visual` | Reversed / highlight |

---

## 19. Per-Project Support

- **Detection:** run `git rev-parse --show-toplevel` to find project root
- **File:** `<git_root>/dooing.json` (configurable filename)
- **Auto-gitignore:** optionally add filename to `.gitignore` (true/false/"prompt")
- **On missing:** `"prompt"` (ask user to create) or `"auto_create"`
- **Context switching:** global and project todos are mutually exclusive at runtime
- **Window title:** shows project directory name when in project mode

---

## 20. Notifications

- **Due date notifications:** on startup and/or when opening the todo window
- **Dedicated due window (`:DooingDue`):** lists all overdue + due-today items
- **Jump to todo:** Enter on a due item opens the main window and jumps to it
- **Dynamic title:** shows count: `" 2 overdue, 3 due today "`

---

## 21. QR Sharing

- TCP HTTP server on port 7283
- `GET /todos` → raw JSON of all todos
- `GET /` → HTML page with QR code pointing to the `/todos` endpoint
- Auto-detects LAN IP, opens browser
- Read-only sharing for mobile access on same network

**Note:** This is a nice-to-have feature, not core. Consider deprioritizing for initial implementation.

---

## 22. Language Recommendations

### Top Pick: Go

**Why:**
- Single binary, zero dependencies — perfect for a CLI/TUI tool
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI framework) is mature, performant, and purpose-built for this
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) for styling (maps perfectly to the highlight system)
- Excellent startup time (<5ms)
- Cross-platform with trivial cross-compilation
- `tmux display-popup -E "dooing"` just works with a single binary

### Runner-Up: Rust

**Why:**
- [Ratatui](https://ratatui.rs/) is excellent for TUI apps
- Best possible performance (though Go is already fast enough)
- Slightly more complex build/development cycle
- Worth it if you want maximum performance or already prefer Rust

### Viable: Zig

If you want absolute minimal binary size and startup time. TUI library ecosystem is less mature.

### Avoid for this use case: Python, Node.js

Startup time penalty, runtime dependency, not ideal for a tool that opens in a tmux popup.

### tmux Integration

```bash
# Add to .tmux.conf or bind to a key:
bind-key t display-popup -E -w 60 -h 22 "dooing"

# Or with project-local todos:
bind-key t display-popup -E -w 60 -h 22 "dooing --project"
```

The app should:
1. Detect terminal size from the environment
2. Render its own bordered window (the tmux popup IS the window frame)
3. Handle input directly (not through tmux keybindings)
4. Exit cleanly on `q` (tmux popup closes automatically when process exits)

---

## Appendix: Original File Map

```
plugin/dooing.vim          → CLI entry point
lua/dooing/init.lua        → Setup, commands, global keymaps
lua/dooing/config.lua      → Default config, merge logic
lua/dooing/state.lua       → Data: CRUD, sort, tags, undo, persist, search, import/export
lua/dooing/server.lua      → HTTP server + QR code
lua/dooing/ui/init.lua     → UI coordinator
lua/dooing/ui/constants.lua → Shared mutable state (window/buffer IDs)
lua/dooing/ui/window.lua   → Window creation, positioning
lua/dooing/ui/rendering.lua → Buffer rendering, fold state
lua/dooing/ui/keymaps.lua  → Keymap registration
lua/dooing/ui/actions.lua  → CRUD UI actions
lua/dooing/ui/components.lua → Help, tags, search, scratchpad windows
lua/dooing/ui/highlights.lua → Highlight groups, priority colors
lua/dooing/ui/utils.lua    → Time formatting, todo line rendering
lua/dooing/ui/calendar.lua → Calendar picker (7 languages)
lua/dooing/ui/due_notification.lua → Due items window
```
