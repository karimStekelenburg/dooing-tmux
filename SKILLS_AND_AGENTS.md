# Recommended Claude Code Skills & Agents

Skills and agents that would allow Claude to implement, test, review, debug, and verify dooing-tmux with minimal human intervention.

---

## Custom Skills to Create

### 1. `go-tui-test` — Visual TUI Snapshot Testing
**Trigger:** After writing or modifying any `ui/` code
**What it does:**
- Runs the app with a predefined `todos.json` fixture
- Captures the rendered output (via `bubbletea/teatest`)
- Compares against golden snapshot files
- Updates snapshots when intentional changes are detected
**Why:** TUI rendering bugs are invisible without visual verification. Bubble Tea's `teatest` package supports programmatic model testing — this skill automates the "does it look right?" check.

### 2. `go-build-verify` — Build + Lint + Test Pipeline
**Trigger:** After any code change (hook on file save in `*.go`)
**What it does:**
- `go build ./...`
- `golangci-lint run`
- `go test ./... -race -count=1`
- Reports first failure with context
**Why:** Catches compilation errors, lint issues, and test failures immediately. The `-race` flag catches concurrency bugs early (relevant for the HTTP server goroutine).

### 3. `go-integration-test` — End-to-End TUI Interaction Test
**Trigger:** After completing any issue, or on-demand
**What it does:**
- Uses `teatest` to simulate full user flows: launch → create todo → toggle → delete → undo → quit
- Verifies the JSON file on disk matches expected state after each operation
- Tests with various terminal sizes (80x24, 60x20, 40x15)
**Why:** Unit tests miss interaction bugs. This catches state machine errors, persistence issues, and rendering problems in combination.

### 4. `dooing-feature-verify` — Issue Acceptance Criteria Checker
**Trigger:** When Claude finishes implementing an issue
**What it does:**
- Reads the GitHub issue's acceptance criteria
- Maps each criterion to a testable assertion
- Runs targeted tests and manual verifications
- Reports pass/fail per criterion
**Why:** Ensures Claude doesn't mark an issue as done when acceptance criteria aren't met.

### 5. `go-benchmark` — Performance Regression Checker
**Trigger:** After changes to sort, render, or persistence code
**What it does:**
- Runs `go test -bench=. -benchmem` on critical paths
- Compares against baseline (stored in `testdata/benchmarks/`)
- Flags regressions > 20%
**Why:** Performance is a stated priority. Sorting 1000 todos, rendering, and JSON serialization must stay fast.

---

## Custom Agents to Create

### 1. `tui-reviewer` — Visual Code Review Agent
**Purpose:** Reviews PRs for TUI-specific issues
**What it checks:**
- Lip Gloss style consistency (are new styles matching the design system?)
- Keybinding conflicts (is a new key already used in another context?)
- Window positioning math (will overlays collide or overflow?)
- Bubble Tea model correctness (does Update return the right Cmd? Is state properly threaded?)
- Terminal compatibility (are Unicode characters used that might not render in basic terminals?)

### 2. `go-test-writer` — Test Generation Agent
**Purpose:** Generates comprehensive tests for new code
**What it does:**
- Reads new/modified Go files
- Generates table-driven tests for all public functions
- Generates `teatest`-based snapshot tests for UI components
- Generates edge case tests (empty input, max values, Unicode, concurrent access)
- Runs the tests it generates to verify they pass
**Why:** Claude often skips edge cases. A dedicated agent focused solely on test quality catches more.

### 3. `issue-implementer` — Autonomous Issue Worker
**Purpose:** Takes a GitHub issue number, implements it end-to-end
**What it does:**
1. Reads the issue description, checklist, acceptance criteria, and dependencies
2. Checks that dependency issues are closed (if not, stops)
3. Creates a feature branch: `feat/<issue-number>-<slug>`
4. Implements the checklist items one by one
5. Writes tests for each item
6. Runs the full test suite
7. Creates a PR linking the issue
8. Self-reviews the PR using `tui-reviewer`
**Why:** This is the main "autopilot" agent. Combined with the skills above, it can implement most issues without human intervention.

### 4. `dependency-checker` — Cross-Issue Consistency Agent
**Purpose:** Verifies integration between features implemented in different issues
**What it does:**
- After any issue is merged, runs integration tests covering feature combinations
- Example: "Create a todo with tags AND priorities AND a due date AND time estimation — does it render correctly? Does sort work?"
- Checks for regressions in previously-completed issues
**Why:** Features built in isolation often break each other. This catches integration issues early.

### 5. `perf-profiler` — Performance Deep-Dive Agent
**Purpose:** Profiles the app under load when performance concerns arise
**What it does:**
- Generates a large `todos.json` (1000+ todos with nested tasks, tags, priorities)
- Profiles startup time, render time, sort time, file I/O
- Uses `go tool pprof` for CPU and memory analysis
- Suggests optimizations if thresholds are exceeded (e.g., startup > 50ms, render > 16ms)
**Why:** Performance is a priority. This agent catches issues before they compound.

---

## Existing Skills to Leverage

| Skill | When to Use |
|---|---|
| `simplify` | After each issue — review for code quality, DRY, efficiency |
| `generate-commit-message` | After each logical unit of work |
| `diff-review` | Before creating PRs — visual before/after comparison |
| `pre-commit-hook-handling` | When pre-commit hooks fail during commits |
| `fact-check` | Verify BREAKDOWN.md stays accurate as implementation diverges |
| `improve-codebase-architecture` | After Phase 2 and Phase 4 — find consolidation opportunities |
| `project-recap` | At each review gate — generate status overview |

---

## Recommended Hooks (settings.json)

```json
{
  "hooks": {
    "post-tool-use": [
      {
        "tool": "Edit",
        "glob": "**/*.go",
        "command": "go build ./... 2>&1 | head -20"
      }
    ],
    "post-tool-use": [
      {
        "tool": "Write",
        "glob": "**/*.go",
        "command": "gofmt -w $FILE && go vet ./..."
      }
    ]
  }
}
```

These hooks ensure every Go file edit is immediately validated — compilation errors surface instantly rather than accumulating.

---

## Autonomy Flow

```
For each issue:
  1. issue-implementer reads issue, checks deps
  2. Creates branch, implements code
  3. go-build-verify runs after each file change (hook)
  4. go-test-writer generates tests
  5. go-tui-test validates rendering
  6. go-integration-test runs full flow
  7. dooing-feature-verify checks acceptance criteria
  8. simplify reviews for quality
  9. Creates PR with diff-review
  10. tui-reviewer does code review
  11. If all green → ready for human review (at gate) or auto-merge (between gates)
```

**Human touchpoints: 4 review gates across 16 issues.**
