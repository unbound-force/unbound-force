# Research: Sub-tool Error Reporting

**Feature**: 036-subtool-error-reporting
**Date**: 2026-06-15

## Research Questions

### RQ-1: What mechanism should surface detailed error output?

**Decision**: Include sub-tool error output inline by default
on failure lines -- no flag needed for P1/P2.

**Rationale**: The sandbox and gateway packages already include
`CombinedOutput()` bytes in error messages inline without any
verbose flag (e.g., `sandbox/podman.go:53`,
`gateway/refresh.go:60`). This is the established project
pattern. Since error output only appears on failure (FR-007
requires no additional output for success), the default inline
approach is sufficient. A `--verbose` flag (P3/FR-006) can be
added later if multi-line output proves too noisy.

**Alternatives considered**:
- `--verbose` flag on `uf init` and `uf setup`: Would require
  plumbing a new flag through Options structs and all callers.
  No precedent exists in the codebase -- no command has
  `--verbose` today. Adds complexity for a problem that inline
  error output already solves.
- `UF_VERBOSE` environment variable: Consistent with the `UF_*`
  env var convention (`UF_PACKAGE_MANAGER`, etc.), but adds a
  hidden switch users must discover. Better as a P3 follow-up
  for full stdout/stderr dumps of successful steps.
- Always-inline with all output: Risk of noisy output when
  sub-tools produce lengthy stderr. Mitigated by truncation
  (see RQ-2).

### RQ-2: What truncation threshold for long error output?

**Decision**: Show the last 10 lines of command output by
default when output exceeds 20 lines. Prefix truncated
output with `... (N lines omitted)`.

**Rationale**: No truncation utility exists in the codebase
today. The 20-line threshold balances diagnostic value with
terminal readability. Most error messages are 1-5 lines;
verbose failures (e.g., Go compiler errors with full stack
traces) can be 50+ lines. Showing the last 10 lines captures
the most diagnostic portion (error summaries and exit messages
typically appear at the end of stderr output).

**Alternatives considered**:
- No truncation (show everything): Risks flooding the terminal
  with 100+ line outputs, burying the summary table.
- Head-based truncation (first N lines): Go compiler errors
  and dependency resolution failures put the root cause at the
  end, not the beginning. Tail-based truncation is more useful.
- Configurable threshold via env var: Over-engineering for the
  initial implementation. Can be added if users report issues.

### RQ-3: How should `subToolResult` be extended?

**Decision**: Add `err error` and `output []byte` fields to
`subToolResult` in `scaffold.go`, mirroring the existing
`stepResult.err` pattern in `setup.go`.

**Rationale**: The `stepResult` struct in `setup.go` already
has an `err error` field (line 138) that is printed by
`printStepResult`. The `subToolResult` struct in `scaffold.go`
is the only result type without error capture. Adding both
`err` and `output` fields:
- `err error` -- carries the `exec.ExitError` (for exit code)
- `output []byte` -- carries the `CombinedOutput()` bytes
  (for the actual stderr/stdout text)

Both are needed because `exec.ExitError.Error()` only returns
`"exit status N"`, which is not the actual error text. The
error text lives in the `[]byte` output from `CombinedOutput()`.

**Alternatives considered**:
- Only add `err error` (like `stepResult`): Insufficient.
  The `err.Error()` from `exec.ExitError` is just
  `"exit status 1"` -- it does not contain the command's
  stderr output. The `[]byte` from `CombinedOutput()` is
  where the actual error text lives.
- Only add `output []byte`: Would lose the structured exit
  code information from `exec.ExitError`. Both are needed
  for FR-004 (report exit code when no stderr).
- Embed output in `detail` string: Would require converting
  `[]byte` to `string` at every call site instead of once
  in the print function. Less clean.

### RQ-4: How should `stepResult` in setup.go be extended?

**Decision**: Add `output []byte` field to `stepResult`.
The `err error` field already exists.

**Rationale**: Setup already wraps errors with context
(`fmt.Errorf("brew install %s: %w", formula, err)`), but
the wrapped error only contains `"exit status N"`. The
actual command output (what brew/dnf/go printed to stderr)
is discarded at every `_, err := opts.ExecCmd(...)` call
site. Adding `output []byte` allows capturing and
displaying this diagnostic text.

### RQ-5: How are package manager alternatives detected?

**Decision**: No new work needed for FR-005. Setup already
mentions alternatives in skip messages with download links.

**Rationale**: Review of `setup.go` shows that skip messages
already include actionable alternatives:
- "Homebrew not available. Download from [URL]"
- "curl|bash install requires --yes flag or interactive
  terminal"
- "Go not available. Install Go or use Homebrew/dnf."

The `resolveMethod()` function (line 207) already iterates
the fallback chain and provides context-appropriate messages.
FR-005 is satisfied by existing behavior. Consistency review
during implementation can improve any outliers.

### RQ-6: How does `exec.ExitError` work with `CombinedOutput()`?

**Decision**: Document for implementer reference.

**Finding**: When `CombinedOutput()` returns an error:
- The error is `*exec.ExitError` when the command runs but
  exits non-zero.
- The `[]byte` output **still contains** stdout+stderr even
  when error is non-nil (Go stdlib documented behavior).
- `exitErr.Error()` returns `"exit status N"` only.
- `exitErr.ExitCode()` returns the integer exit code.
- The `cmdRecorder` test stubs already support returning
  both output and error simultaneously.

This means the fix is straightforward: change `_, err :=`
to `out, err :=` at ExecCmd call sites, and pass `out` to
the result struct.

## Coverage Strategy

**Unit tests**: All changes are in `scaffold.go` and
`setup.go`. Both have comprehensive test files with
`cmdRecorder` / `scaffoldCmdRecorder` stubs that already
test failure scenarios.

**Testing approach**:
- Extend existing failure tests to assert that the error
  output appears in the rendered summary (check
  `buf.String()` for sub-tool stderr text).
- Add new tests for truncation behavior (output > 20
  lines).
- Add new tests for exit-code-only fallback (FR-004).
- Use table-driven tests consistent with existing patterns.
- No new test infrastructure needed -- existing
  `cmdRecorder` already supports returning both output
  and error.

**Coverage targets**: Existing test files cover ~30+
failure scenarios across both packages. New tests should
cover:
1. Error output propagation (scaffold: 5 call sites)
2. Error output propagation (setup: ~25 call sites)
3. Truncation with output > 20 lines
4. Exit code display when output is empty
5. No output change for successful steps
