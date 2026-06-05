## Why

`uf init` runs five sub-tool initializations (dewey, replicator,
specify, openspec, gaze) sequentially via blocking `exec.Command`
calls. Dewey indexing alone can take 30-60+ seconds due to web
crawling, source fetching, and embedding generation. The remaining
four tools add additional wall-clock time despite being entirely
independent of each other and of Dewey.

Users experience `uf init` as "hanging" with no feedback while
these sequential operations complete.

## What Changes

Refactor `initSubTools()` in `internal/scaffold/scaffold.go` to
run independent sub-tool initializations concurrently using
`sync.WaitGroup` + `sync.Mutex`. Dewey init/index is the
slowest step and should not block the other tools. The four
non-Dewey tools (replicator, specify, openspec, gaze) are
independent of each other and of Dewey.

Additionally, `initSubTools()` now reads the project config
(`setup.tools.<name>.method: skip`) before launching any
goroutines. Tools configured with `method: skip` are excluded
entirely — no goroutine is spawned, no binary is looked up,
and no results are produced. This uses the existing
`config.ToolConfig.Method` field already defined in the
config package.

## Capabilities

### New Capabilities
- `concurrent-subtool-init`: Sub-tools that have no
  interdependencies run in parallel via `sync.WaitGroup`
- `config-tool-skip`: Tools with `setup.tools.<name>.method:
  skip` in `.uf/config.yaml` are excluded from initialization

### Modified Capabilities
- `initSubTools`: Returns the same `[]subToolResult` but
  completes faster by overlapping independent work

### Removed Capabilities
- None

## Impact

- `internal/scaffold/scaffold.go`: `initSubTools()` function
  refactored for concurrent execution and config-based skip
- `internal/scaffold/scaffold_test.go`: Tests updated to
  verify concurrent behavior, result aggregation, and
  config-based tool skipping
- No API changes, no new dependencies (uses stdlib `sync`)

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is internal to the scaffold engine. It does not
affect artifact-based communication between heroes.

### II. Composability First

**Assessment**: PASS

Each sub-tool remains independently installable and callable.
The change only affects the orchestration of their init calls,
not their interfaces. `LookPath` checks are preserved -- missing
tools are still skipped gracefully.

### III. Observable Quality

**Assessment**: PASS

Sub-tool results are still collected and returned as
`[]subToolResult`. Logging output from each tool is captured.
The change preserves all existing observability.

### IV. Testability

**Assessment**: PASS

`initSubTools` already uses injected `ExecCmd` and `LookPath`
functions on the `Options` struct, enabling test doubles. The
concurrent execution can be tested by verifying that results
from all tools are present regardless of execution order.
