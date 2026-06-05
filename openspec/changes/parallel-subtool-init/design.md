## Context

`initSubTools()` in `internal/scaffold/scaffold.go` runs five
sub-tool initializations sequentially: dewey, replicator,
specify, openspec, and gaze. Each uses `opts.ExecCmd()` which
calls `exec.Command(...).CombinedOutput()` -- blocking until
the process completes.

Dewey is the slowest (30-60+ seconds for initial indexing due
to web crawling, source fetching, and embedding generation).
The remaining four tools each take 1-5 seconds but add up when
run serially.

Per the proposal's constitution alignment, this change is
internal to the scaffold engine and maintains composability
(each tool remains independently callable) and testability
(injected `ExecCmd`/`LookPath` functions enable test doubles).

## Goals / Non-Goals

### Goals
- Run independent sub-tool initializations concurrently to
  reduce total wall-clock time
- Preserve existing error handling semantics (non-fatal
  per-tool failures)
- Maintain result collection (`[]subToolResult`) with same
  content regardless of execution order
- Thread-safe result aggregation

### Non-Goals
- Optimizing Dewey's internal indexing pipeline (separate
  change in the dewey repo)
- Adding progress bars or streaming output (could be a
  follow-up)
- Running Dewey init in the background (fire-and-forget)
  -- all tools must complete before `initSubTools` returns
- Changing the `Options` struct API

## Decisions

### D1: Use `sync.WaitGroup` + mutex, not `errgroup`

`errgroup` cancels remaining goroutines on first error. Sub-tool
failures are non-fatal -- we want all tools to attempt init
regardless of individual failures. Use `sync.WaitGroup` for
coordination and `sync.Mutex` for thread-safe result collection.

### D2: Two concurrency groups

Group the tools into two independent concurrent groups based on
their internal dependencies:

- **Group A (Dewey)**: `dewey init` -> `generateDeweySources`
  -> `dewey index` (sequential within this group)
- **Group B (others)**: replicator, specify, openspec, gaze
  (all independent, run concurrently with each other and with
  Group A)

Rationale: Dewey has an internal dependency chain (init must
precede source generation which must precede indexing). The
other four tools have no dependencies on each other or on Dewey.

### D3: Capture stdout/stderr per tool

Each goroutine captures its own tool's output via the existing
`ExecCmd` mechanism. Results are appended to the shared
`[]subToolResult` slice under a mutex. Output interleaving is
not a concern since `ExecCmd` uses `CombinedOutput()` (already
captures per-process).

### D4: configureOpencodeJSON runs after all tools complete

`configureOpencodeJSON()` (line 1424) depends on tool
availability checks. It MUST run after all concurrent groups
complete. This is enforced by placing it after `wg.Wait()`.

## Risks / Trade-offs

- **Risk**: Concurrent file writes if two tools write to the
  same file. **Mitigation**: Each tool writes to its own
  directory (`.uf/dewey/`, `.specify/`, `.openspec/`,
  `.uf/gaze/`). No overlap.
- **Risk**: Error output harder to read if multiple tools fail
  simultaneously. **Mitigation**: Each failure is captured in
  its own `subToolResult` with tool name. Acceptable trade-off
  for reduced wall-clock time.
- **Trade-off**: Slightly more complex code in `initSubTools`.
  Justified by the user-facing latency improvement.
