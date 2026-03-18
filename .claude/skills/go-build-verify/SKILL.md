---
name: go-build-verify
description: "Run the Go build/lint/test verification pipeline. Use this skill after ANY edit to a .go file, when the user says 'verify', 'check build', 'run tests', 'does it compile', 'does it pass', 'check if it works', or after completing a checklist item during issue implementation. Also trigger this when you've just finished writing Go code and need to confirm it works. When in doubt about whether to verify — verify."
---

# Go Build Verify

Run build, lint, and tests in sequence. Stop at the first failure so it can be fixed immediately.

## Why short-circuit matters

Running tests when the build is broken wastes time and produces confusing output (import errors, missing symbols). Linting a broken build is equally useless. Always go in order: compile first, lint second, test third.

## Steps

Run these from the project root (`/Users/krm_xel/dev/dooing-tmux`). Stop at the first step that fails.

### 1. Build

```bash
go build ./... 2>&1
```

If this fails, report the compiler errors and stop. Compiler errors are the highest priority — nothing else matters until the code compiles.

### 2. Lint (if golangci-lint is available)

```bash
cd /Users/krm_xel/dev/dooing-tmux && golangci-lint run ./... 2>&1
```

If `golangci-lint` is not installed, skip this step silently (do not install it or ask the user about it). If it runs and finds issues, report them and stop.

### 3. Test

```bash
cd /Users/krm_xel/dev/dooing-tmux && go test ./... -race -count=1 2>&1
```

Why these flags:
- `-race` detects data races at runtime. This project will have an HTTP server goroutine (QR sharing feature), so race conditions are a real risk.
- `-count=1` disables test caching so results are always fresh. Cached passes can hide regressions introduced by recent edits.

If tests fail, report the failing test name, the assertion that failed, and the relevant output.

## Reporting

- **On failure:** Show the raw output from the failing step. Include enough context (10-20 lines around the error) for immediate diagnosis. State which step failed (build/lint/test) so priority is clear.
- **On success:** Report concisely: "Build, lint, and tests pass." No need to dump passing output.
