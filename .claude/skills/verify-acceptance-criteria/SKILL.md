---
name: verify-acceptance-criteria
description: Verifies whether a GitHub issue's acceptance criteria have been met by the current codebase. Use this skill whenever the user asks to check acceptance criteria, verify an issue is done, asks "are we done with #N", "does this meet the requirements", "check if issue N is complete", or after finishing implementation of a GitHub issue. Also use when the user says "verify issue", "check criteria", or "acceptance check".
---

# Verify Acceptance Criteria

This skill checks whether a GitHub issue's acceptance criteria are satisfied by the current code, tests, and build state. It produces a per-criterion pass/fail report with evidence.

## Workflow

### Step 1: Fetch the issue

Run `gh issue view <NUMBER> --repo karimStekelenburg/dooing-tmux` to get the issue body. If the user provides a number, use it directly. If they say something like "the current issue" or "this issue", check the current branch name for an issue number (e.g., `feat/7-tag-system` implies issue #7).

### Step 2: Extract acceptance criteria

Look for a section titled "Acceptance Criteria" (or similar: "Done when", "Definition of Done", "Requirements", "Checklist"). Criteria are typically in checkbox format:

```
- [ ] Todos can be created with inline #tags
- [ ] Tags are extracted and stored in the category field
- [ ] All tests pass
```

If no explicit acceptance criteria section exists, extract testable requirements from the issue description — bullet points, numbered lists, or imperative statements that describe expected behavior.

### Step 3: Classify each criterion

For each criterion, determine verification strategy:

| Pattern | Strategy | How to verify |
|---|---|---|
| Code/feature exists | **grep** | Search for relevant types, functions, or constants in the codebase |
| Tests pass | **run** | Execute `go test ./... -run <pattern>` or the full suite |
| Build succeeds | **run** | Execute `go build ./...` |
| Lint passes | **run** | Execute `golangci-lint run` if available |
| Data persists to disk | **test** | Verify a test exists that checks file I/O, then run it |
| UI renders correctly | **code+test** | Confirm rendering code exists and snapshot/teatest tests pass |
| Keybinding works | **code** | Verify the key is registered in the keymap and handler exists |
| Subjective/UX quality | **flag** | Cannot be verified automatically — flag for human review |

### Step 4: Verify each criterion

For each criterion, collect evidence:

1. **grep/glob** the codebase to find relevant code
2. **Run** relevant tests or commands
3. Record the result: PASS (with evidence), FAIL (with what's missing), or NEEDS_HUMAN_REVIEW (with explanation)

Be thorough but efficient — run the full test suite once rather than individual tests repeatedly.

### Step 5: Report results

Present results in this format:

```
## Acceptance Criteria Verification: Issue #<N>

**Title:** <issue title>
**Verdict:** <ALL CRITERIA MET / N of M criteria remaining / NEEDS HUMAN REVIEW>

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| 1 | Todos can be created with inline #tags | PASS | `todo.go:CreateTodo()` extracts tags via regex; `todo_test.go:TestCreateWithTags` passes |
| 2 | Calendar renders month grid | NEEDS HUMAN REVIEW | Code exists in `calendar.go` but visual correctness requires human inspection |
| 3 | All tests pass | FAIL | 2 failures in `sort_test.go` — see output below |

### Failures (details)
<paste relevant test output or missing code details>

### Items Needing Human Review
<list subjective criteria with context on what to look at>
```

## Important guidelines

- Always run `go build ./...` as a baseline check before anything else — if it doesn't compile, nothing else matters.
- When a criterion is ambiguous, interpret it generously but note your interpretation.
- If tests don't exist for a criterion, that's a FAIL — the criterion asks for verified behavior, not just code presence. Note that tests should be written.
- Don't fabricate evidence. If you can't find the relevant code, say so.
- For criteria about "all tests pass", run the full suite and report the actual result, including any failures.
- Group related criteria when they can be verified by the same test run to avoid redundant work.
