# Data Model: Sub-tool Error Reporting

**Feature**: 036-subtool-error-reporting
**Date**: 2026-06-15

## Entity Changes

### subToolResult (scaffold.go, line 476)

**Current fields**:

| Field  | Type   | Description                                    |
|--------|--------|------------------------------------------------|
| name   | string | Tool/step name (e.g., ".specify/", "dewey index") |
| action | string | Status vocabulary: initialized, completed, failed, skipped, created, configured, already configured, overwritten, error, dry-run |
| detail | string | Human-readable summary message                 |

**Proposed additions**:

| Field  | Type   | Description                                    |
|--------|--------|------------------------------------------------|
| err    | error  | Underlying error from ExecCmd (exec.ExitError) |
| output | []byte | Captured stdout+stderr from CombinedOutput()   |

**Rationale**: `err` carries the structured exit code
(`exec.ExitError.ExitCode()`). `output` carries the actual
error text that the sub-tool printed. Both are needed because
`err.Error()` only returns `"exit status N"` without the
diagnostic text.

### stepResult (setup.go, line 133)

**Current fields**:

| Field  | Type   | Description                                    |
|--------|--------|------------------------------------------------|
| name   | string | Tool name (e.g., "gaze", "ollama")             |
| action | string | installed, already installed, skipped, failed   |
| detail | string | Human-readable summary with remediation hints   |
| err    | error  | Wrapped error from ExecCmd                      |

**Proposed additions**:

| Field  | Type   | Description                                    |
|--------|--------|------------------------------------------------|
| output | []byte | Captured stdout+stderr from CombinedOutput()   |

**Rationale**: The `err` field already exists but only
contains `"exit status N"` wrapped with context. The actual
command output (what brew/dnf/go printed) is discarded.
Adding `output` closes this gap.

## New Helper

### truncateOutput

**Purpose**: Truncate long command output for display,
keeping the most diagnostic portion (tail).

**Signature**:
```
Input:  output []byte, maxLines int
Output: string
```

**Behavior**:
- If output has <= maxLines lines: return full output as
  string (trimmed).
- If output has > maxLines lines: return last
  (maxLines/2) lines (integer division), prefixed with
  `"... (N lines omitted)\n"`.
- Empty output returns empty string.
- Default: `maxLines=20`, showing the last 10 lines.

**Location**: Defined in the package where it is used.
If both packages need it, define in each (small function,
not worth a shared package for 2 call sites). If either
copy is modified, both MUST be updated together.

## State Transitions

No state transitions. The `action` field vocabulary is
unchanged. The new fields add diagnostic context to
existing failure states without altering the state
machine.

## Relationships

- `subToolResult.err` and `subToolResult.output` are
  populated together from the same `ExecCmd()` call.
- `stepResult.output` supplements the existing
  `stepResult.err` field.
- `printSummary` (scaffold) and `printStepResult` (setup)
  consume the new fields for display.
