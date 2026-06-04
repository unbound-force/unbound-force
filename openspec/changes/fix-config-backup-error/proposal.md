## Why

`internal/config/init.go` writes a backup of the existing
config file before overwriting it with an updated version.
The backup write error is silently discarded with `_ =`,
meaning a failure (disk full, permission denied) does not
prevent the subsequent overwrite of the live config. The
user loses their configuration with no diagnostic and no
recovery path.

This violates CS-006 [MUST]: "Errors MUST NOT be silently
swallowed."

Issue: https://github.com/unbound-force/unbound-force/issues/237

## What Changes

- `internal/config/init.go`: replace `_ = opts.WriteFile(...)`
  on the backup write with a proper error check that aborts
  before the live config is touched.
- `internal/config/init_test.go`: add a regression test
  (`TestInitFile_BackupWriteFailureAbortsUpdate`) that
  injects a failing `WriteFile` stub and asserts the error
  is surfaced and the original config is unmodified.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `InitFile`: now returns an error wrapping `"write backup
  config"` when the backup write fails, and does not proceed
  to overwrite the live config file.

### Removed Capabilities
- None.

## Impact

- `internal/config/init.go` — one-line change at line 80.
- `internal/config/init_test.go` — one new test function.
- No API or CLI surface change; `InitFile` already returned
  `error`; callers already handle it.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is internal to a single function with no
artifact-based communication surface. It does not affect
how heroes exchange artifacts.

### II. Composability First

**Assessment**: N/A

`InitFile` is an internal utility. No public interfaces,
CLI flags, or external dependencies are added or changed.

### III. Observable Quality

**Assessment**: PASS

Before this fix, a backup write failure produced no
observable output — the error was swallowed. After this
fix, the error is returned to the caller and propagated
up to the CLI layer, where it is printed and the process
exits non-zero. Failures are now observable.

### IV. Testability

**Assessment**: PASS

The fix is verified by a regression test that uses the
existing `WriteFile` injection point (`InitOptions.WriteFile`)
to simulate a backup failure without touching the
filesystem. The component remains fully testable in
isolation.
