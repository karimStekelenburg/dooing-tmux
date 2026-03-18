# Implementation Plan

## Dependency Graph

```
#1 Scaffolding
 └─► #2 Data Model + Config
      ├─► #3 Core UI + Navigation ─► #4 Create/Edit ──┐
      │                            ─► #5 Toggle/Delete ┤
      │                                                ▼
      │                                         #6 Sort + Help
      │                                          ├─► #7  Tags
      │                                          ├─► #8  Priorities
      │                                          ├─► #9  Nested Tasks
      │                                          ├─► #10 Calendar ──► #15 Due Notifications
      │                                          ├─► #11 Time + Search
      │                                          ├─► #12 Scratchpad
      │                                          └─► #14 Import/Export
      ├─► #13 Per-project + tmux
      └─► #16 QR Sharing
```

## Phases & Review Gates

| Phase | Issues | Milestone | Human Review? |
|-------|--------|-----------|---------------|
| 1 | #1 → #2 → #3 | Working TUI with rendered todos | ✅ Yes |
| 2 | #4 ∥ #5 → #6 | Full CRUD + undo + sort | ✅ Yes |
| 3 | #7 ∥ #8 | Tags + priorities | ✅ Yes |
| 4 | #9 ∥ #10 ∥ #11 ∥ #12 | Feature-complete core | No (CI gates) |
| 5 | #13 ∥ #14 ∥ #15 ∥ #16 | Production-ready | ✅ Yes (final) |

**Total human review gates: 4**

## Parallelization

- Phase 2: #4 and #5 can be parallel branches
- Phase 3: #7 and #8 can be parallel branches
- Phase 4: ALL FOUR issues (#9, #10, #11, #12) can be parallel branches
- Phase 5: ALL FOUR issues (#13, #14, #15, #16) can be parallel (except #15 needs #10)
