<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Config-Based Tool Skipping

- [x] 1.1 Add `shouldSkipTool()` to `initSubTools()` that
  reads `setup.tools.<name>.method: skip` from project
  config via `config.Load()`. All five tools (dewey,
  replicator, specify, openspec, gaze) check this before
  `LookPath`. Implements FR-006.
  **Files**: `internal/scaffold/scaffold.go`

- [x] 1.2 Add tests for config-based skipping: single tool
  skip, skip overrides force, replicator skip, all tools
  skipped. Implements FR-006 scenarios.
  **Files**: `internal/scaffold/scaffold_test.go`

## 2. Refactor initSubTools for Concurrency

- [x] 2.1 Refactor `initSubTools()` in
  `internal/scaffold/scaffold.go` to use `sync.WaitGroup`
  and `sync.Mutex` for concurrent sub-tool execution.
  Group A (dewey init -> generateDeweySources ->
  dewey index) runs in one goroutine. Group B tools
  (replicator, specify, openspec, gaze) each run in
  their own goroutine. `configureOpencodeJSON()` runs
  after `wg.Wait()`. Implements FR-001, FR-002, FR-003,
  FR-004, FR-005.
  **Files**: `internal/scaffold/scaffold.go`

## 3. Tests

- [x] 3.1 Add test(s) verifying concurrent execution
  behavior: all tool results are collected, missing
  tools are skipped, one tool failure does not block
  others, `configureOpencodeJSON` runs after all tools
  complete. Use injected `ExecCmd`/`LookPath` test
  doubles.
  **Files**: `internal/scaffold/scaffold_test.go`

## 4. Verification

- [x] 4.1 Run `make check` (lint + test + build) and
  confirm all pass with `-race -count=1`
- [x] 4.2 Verify constitution alignment: composability
  (each tool independently callable), testability
  (injected functions enable test doubles), observable
  quality (results still collected and returned)
