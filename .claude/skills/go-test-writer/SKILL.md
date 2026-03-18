---
name: go-test-writer
description: Generate comprehensive Go tests for new or modified code, including table-driven tests, teatest snapshot tests, edge cases, and state machine tests. Use this skill whenever the user says "write tests", "add tests", "test this", "generate tests", "test coverage", "edge cases", asks for tests after implementing a feature, or mentions testing any Go code in this project. Even if the user just says "ok now test it" or "make sure this works", use this skill.
---

# Go Test Writer

Generate thorough, idiomatic Go tests for dooing-tmux code. The key principle: **read the code first, then write tests**. Never generate tests blind.

## Workflow

1. **Read the target file(s).** Understand every public function's signature, types, return values, and side effects. Also read any types/interfaces the functions depend on.

2. **Check go.mod** for available test dependencies (testify, teatest, etc.). Use what's already there — don't add new dependencies without asking.

3. **Determine test categories** based on what the code does:

   - **Pure logic** (sorting, scoring, parsing, tag extraction) — table-driven tests
   - **UI components** (Bubble Tea models) — teatest snapshot tests
   - **State mutations** (todo CRUD, toggle cycle) — state machine tests
   - **Persistence** (JSON read/write) — round-trip and migration tests
   - **All categories** — edge case tests

4. **Write the test file** as `<name>_test.go` alongside the source file, in the same package.

5. **Run `go test ./... -race -count=1`** to verify tests pass.

6. **If tests fail, fix them.** Read the error output, fix the test (not the source code unless there's a genuine bug), and rerun. Iterate until green.

## Test Patterns

### Table-Driven Tests

The standard pattern for this project. Every public function gets one.

```go
func TestParseTimeEstimation(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  float64
        err   bool
    }{
        {"minutes", "15m", 0.25, false},
        {"hours", "2h", 2.0, false},
        {"days", "1d", 8.0, false},
        {"weeks", "0.5w", 20.0, false},
        {"empty string", "", 0, true},
        {"invalid unit", "5x", 0, true},
        {"negative", "-2h", 0, true},
        {"zero", "0h", 0, false},
        {"unicode input", "２h", 0, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseTimeEstimation(tt.input)
            if tt.err {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.InDelta(t, tt.want, got, 0.001)
        })
    }
}
```

### Edge Cases to Always Include

For every function, think about these inputs:
- **Zero/nil values**: empty string, nil slice, nil pointer, zero int
- **Boundary values**: max int, very long strings (1000+ chars)
- **Unicode**: emoji in todo text, CJK characters in tags, RTL text
- **Special characters**: `#` in non-tag position, newlines in text, quotes
- **Concurrent access**: if the function touches shared state, test with goroutines

### Teatest Snapshot Tests (UI Components)

For any Bubble Tea model, test that it renders correctly given known state.

```go
func TestMainModelRender(t *testing.T) {
    m := NewMainModel(WithTodos(fixtureTodos()))
    tm := teatest.NewModel(t, m, teatest.WithInitialTermSize(80, 24))
    tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{}})
    out := tm.FinalOutput(t)
    teatest.RequireEqualOutput(t, out) // compares against .golden file
}
```

Golden files go in `testdata/` next to the test file. On first run, pass `-update` flag to create them.

### State Machine Tests (Todo Toggle Cycle)

The toggle cycle is a core invariant: pending -> in_progress -> done -> pending.

```go
func TestToggleCycle(t *testing.T) {
    todo := NewTodo("test task")

    // pending -> in_progress
    todo.Toggle()
    assert.False(t, todo.Done)
    assert.True(t, todo.InProgress)

    // in_progress -> done
    todo.Toggle()
    assert.True(t, todo.Done)
    assert.False(t, todo.InProgress)
    assert.NotNil(t, todo.CompletedAt)

    // done -> pending
    todo.Toggle()
    assert.False(t, todo.Done)
    assert.False(t, todo.InProgress)
    assert.Nil(t, todo.CompletedAt)
}
```

### JSON Round-Trip Tests (Persistence)

Serialize, deserialize, compare. Test with edge-case data.

```go
func TestJSONRoundTrip(t *testing.T) {
    todos := []Todo{
        {Text: "normal task", Done: false},
        {Text: "task with #tag and emoji 🎉", Done: true, CompletedAt: timePtr(now)},
        {Text: "nested task", ParentID: strPtr("123"), Depth: 2},
        {Text: "", Done: false}, // empty text edge case
    }
    data, err := json.Marshal(todos)
    assert.NoError(t, err)

    var got []Todo
    err = json.Unmarshal(data, &got)
    assert.NoError(t, err)
    assert.Equal(t, todos, got)
}
```

### Migration Tests

Test that old JSON formats (missing fields) load correctly with defaults backfilled.

```go
func TestMigrationBackfillsFields(t *testing.T) {
    // Old format: no id, no parent_id, no depth
    oldJSON := `[{"text":"old todo","done":false}]`
    todos, err := LoadFromJSON([]byte(oldJSON))
    assert.NoError(t, err)
    assert.NotEmpty(t, todos[0].ID)    // backfilled
    assert.Equal(t, 0, todos[0].Depth) // default
    assert.Nil(t, todos[0].ParentID)   // default nil
}
```

## What NOT to Do

- Don't test private functions directly — test them through public API
- Don't mock what you can construct — prefer real structs over mocks when possible
- Don't write tests that depend on wall-clock time — inject time or use relative assertions
- Don't generate tests without reading the source first — you'll get signatures wrong
