---
name: go-benchmark
description: Run Go benchmarks and detect performance regressions in dooing-tmux. Use this skill whenever the user mentions "benchmark", "check performance", "perf regression", "is this fast enough", "how fast is", or "performance test". Also use it after changes to sorting, rendering, or persistence code, and before creating PRs for performance-sensitive issues. Even if the user just says something like "make sure I didn't slow things down", this skill applies.
---

# Go Benchmark — Performance Regression Checker

This skill runs Go benchmarks on dooing-tmux's critical paths, compares results against a stored baseline, and flags regressions. The goal is to keep the app fast enough for 60fps rendering in a tmux popup.

## Performance Budgets

These are the hard limits. Any result exceeding these is a regression worth investigating:

| Operation | Budget | Rationale |
|---|---|---|
| Sort 1000 todos | <5ms | Must not block UI thread |
| Render 100 todos (View()) | <16ms | 60fps frame budget |
| JSON load 1000 todos | <10ms | Startup speed |
| Startup to first render | <50ms | Perceived instant |

## Workflow

### Step 1: Run benchmarks

Run benchmarks with 3 iterations for statistical reliability:

```bash
go test -bench=. -benchmem -count=3 -timeout=120s ./... 2>&1 | tee /tmp/dooing-bench-latest.txt
```

If only specific packages changed, scope the benchmarks:

- Sort/priority changes: `go test -bench=. -benchmem -count=3 ./internal/state/...`
- Rendering changes: `go test -bench=. -benchmem -count=3 ./internal/ui/...`
- Persistence changes: `go test -bench=. -benchmem -count=3 ./internal/persistence/...`

If no benchmark functions exist yet in the relevant package, tell the user and offer to create them. Use these patterns:

```go
func BenchmarkSortTodos(b *testing.B) {
    todos := generateTodos(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        SortTodos(todos)
    }
}

func BenchmarkRenderView(b *testing.B) {
    m := NewModel(testConfig, generateTodos(100))
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = m.View()
    }
}

func BenchmarkJSONLoad(b *testing.B) {
    data := generateTodosJSON(1000)
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        LoadTodos(data)
    }
}
```

The `generateTodos` helper should create realistic todos with tags, priorities, due dates, nested tasks, and notes to represent real-world data.

### Step 2: Compare against baseline

Check if a baseline exists:

```bash
ls testdata/benchmarks/baseline.txt 2>/dev/null
```

**If baseline exists**, compare using `benchstat` (preferred) or manual parsing:

```bash
# If benchstat is installed:
benchstat testdata/benchmarks/baseline.txt /tmp/dooing-bench-latest.txt

# If not installed, install it:
go install golang.org/x/perf/cmd/benchstat@latest
benchstat testdata/benchmarks/baseline.txt /tmp/dooing-bench-latest.txt
```

If `benchstat` cannot be installed, parse the output manually and compare ns/op, B/op, and allocs/op values between baseline and current.

### Step 3: Evaluate results

Flag as a regression if:
- **Time (ns/op)** increased by >20%
- **Allocations (allocs/op)** increased by >50%
- **Memory (B/op)** increased by >50%
- Any benchmark exceeds the performance budgets listed above

### Step 4: Report results

Present a clear summary table:

```
## Benchmark Results

| Benchmark | Baseline | Current | Delta | Status |
|---|---|---|---|---|
| BenchmarkSortTodos-8 | 2.1ms | 2.3ms | +9.5% | PASS |
| BenchmarkRenderView-8 | 8.4ms | 14.2ms | +69% | FAIL - exceeds 20% threshold |
| BenchmarkJSONLoad-8 | 3.1ms | 3.0ms | -3.2% | PASS |

Allocation changes:
| Benchmark | Baseline allocs | Current allocs | Delta | Status |
|---|---|---|---|---|
| BenchmarkSortTodos-8 | 15 | 15 | 0% | PASS |
| BenchmarkRenderView-8 | 230 | 890 | +287% | FAIL - exceeds 50% threshold |
```

If there are regressions, suggest specific investigation steps (e.g., "Run `go test -bench=BenchmarkRenderView -cpuprofile=cpu.prof` and inspect with `go tool pprof`").

If there is no baseline yet, just report the absolute numbers and whether they meet the performance budgets.

### Step 5: Update baseline (only when asked)

Only update the baseline when the user explicitly says to, or passes `--update-baseline`:

```bash
mkdir -p testdata/benchmarks
cp /tmp/dooing-bench-latest.txt testdata/benchmarks/baseline.txt
```

Also save metadata:

```bash
echo "# Baseline captured: $(date -u +%Y-%m-%dT%H:%M:%SZ)" > testdata/benchmarks/baseline_info.txt
echo "# Git ref: $(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" >> testdata/benchmarks/baseline_info.txt
echo "# Go version: $(go version)" >> testdata/benchmarks/baseline_info.txt
```

## Important Notes

- Never update the baseline without the user's explicit approval. Silently updating would hide regressions.
- If benchmarks don't exist for the changed code, say so clearly rather than silently skipping. Offer to write them.
- When reporting, always include both time and allocation metrics. Allocation growth often predicts future GC pressure even when wall-clock time looks fine.
- For flaky results (high variance across runs), increase `-count` to 5 or 10 and note the variance.
