---
name: go-tui-test
description: Run visual TUI snapshot tests using Bubble Tea's teatest package to verify rendering correctness. Use this skill after modifying any file in internal/ui/, when the user says "check rendering", "snapshot test", "visual test", "does it look right", or after implementing any UI-related issue (#3, #7, #8, #9, #10, #11, #12). Also use when debugging layout problems, verifying style changes, or ensuring a refactor didn't break the visual output.
---

# Visual TUI Snapshot Testing

This skill runs golden-file snapshot tests against the TUI's rendered output to catch rendering regressions. TUI bugs are invisible without visual verification — a passing `go build` tells you nothing about whether the UI actually looks correct.

## How it works

1. Load a deterministic `testdata/todos.json` fixture into the model
2. Use `teatest` to programmatically drive the Bubble Tea model through key states
3. Capture the `View()` string output at each state
4. Compare against golden files in `testdata/snapshots/`
5. If a golden file is missing, create it (first run bootstrapping)
6. If a golden file differs, show the diff and ask the user whether to update

## Step-by-step

### 1. Check for existing test infrastructure

Look for snapshot test files:

```
testdata/
  todos.json              # fixture data
  snapshots/
    *.golden              # golden snapshot files
internal/ui/
  *_test.go               # test files using teatest
```

If none exist yet, create them following the patterns below.

### 2. Fixture data

The fixture at `testdata/todos.json` should be deterministic — fixed timestamps, fixed IDs, a representative mix of todo states. Include:
- 2-3 pending todos (one with tags, one with due date, one with priority)
- 1 in-progress todo
- 2 done todos
- 1 nested todo (if nesting is implemented)
- At least 2 different tags

Use fixed `created_at` timestamps (e.g., `1704067200` = 2024-01-01) so relative time displays are predictable. When relative time would change between runs, either mock the clock or strip relative timestamps before comparison.

### 3. Write snapshot tests

Each test follows this pattern:

```go
func TestMainViewPopulated(t *testing.T) {
    todos := loadFixture(t, "testdata/todos.json")
    m := ui.NewModel(testConfig(), todos)

    // teatest creates a test model with a fixed terminal size
    tm := teatest.NewModel(t, m, teatest.WithInitialTermSize(55, 20))

    // Drive to desired state (if needed)
    // tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})

    // Wait for the model to settle
    out := tm.FinalModel(t, teatest.WithFinalTimeout(time.Second))

    // Compare against golden file
    golden.RequireEqual(t, []byte(out.View()))
}
```

The `golden` package (from `github.com/charmbracelet/x/exp/golden`) handles the golden file comparison. Run tests with `-update` flag to regenerate snapshots: `go test ./internal/ui/ -update`.

### 4. States to test

Cover these states with individual test functions. Each produces a separate `.golden` file:

| Test name | State | Key actions |
|---|---|---|
| `TestMainViewEmpty` | Empty todo list | None — just render with no todos |
| `TestMainViewPopulated` | Normal populated list | Load fixture, render |
| `TestMainViewFiltered` | Filtered by tag | Send key to open tags, select a tag |
| `TestHelpOverlay` | Help window visible | Send `?` key |
| `TestCalendarWidget` | Calendar picker open | Send `H` key on a todo |
| `TestPrioritySelector` | Priority selector open | Send `p` key on a todo |

### 5. Run the tests

```bash
# Run snapshot tests
go test ./internal/ui/ -run "Test.*View|Test.*Overlay|Test.*Widget|Test.*Selector" -v

# Update golden files after intentional changes
go test ./internal/ui/ -run "Test.*View|Test.*Overlay|Test.*Widget|Test.*Selector" -update
```

### 6. Interpret results

- **PASS with no golden file changes**: Rendering is stable, no regressions.
- **FAIL with diff output**: Something changed. Read the diff carefully.
  - If the change is intentional (you just modified the UI): update with `-update` flag.
  - If the change is unintentional: you have a regression. Fix it before proceeding.
- **New golden file created**: First run for this test. Review the `.golden` file contents to confirm they look correct.

After updating golden files, always show the user what changed so they can confirm the new rendering is correct.

### 7. Handling time-dependent output

Relative timestamps (`@3h ago`) change between runs. Two approaches:

**Option A (preferred):** Mock the time source. Pass a `func() time.Time` into the model and use a fixed time in tests.

**Option B:** Strip or normalize time-dependent parts before comparison:

```go
re := regexp.MustCompile(`@\S+ ago|just now`)
normalized := re.ReplaceAllString(out.View(), "@TIME")
golden.RequireEqual(t, []byte(normalized))
```

### 8. Terminal size

Always use `teatest.WithInitialTermSize(55, 20)` to match the default window size from the spec. For responsive layout testing, add additional tests at other sizes (e.g., 40x15, 80x24) with separate golden files.

## Snapshot file organization

```
testdata/snapshots/
  TestMainViewEmpty.golden
  TestMainViewPopulated.golden
  TestMainViewFiltered.golden
  TestHelpOverlay.golden
  TestCalendarWidget.golden
  TestPrioritySelector.golden
```

The `golden` package automatically names files after the test function, so the file organization happens naturally.

## Common pitfalls

- **Forgetting terminal size**: Without `WithInitialTermSize`, the output depends on the test runner's terminal, making snapshots flaky.
- **Non-deterministic data**: Random IDs or timestamps in fixtures cause spurious diffs. Always use fixed values.
- **ANSI escape codes**: Golden files include styling codes. If you change a Lip Gloss style, all affected snapshots will diff — this is expected and correct.
- **Race conditions**: If the model uses `Cmd`s that resolve asynchronously, use `teatest.WaitFor` to wait for specific output before capturing the snapshot.
