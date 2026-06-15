# Quickstart: Sub-tool Error Reporting

**Feature**: 036-subtool-error-reporting
**Date**: 2026-06-15

## What Changes

`uf init` and `uf setup` failure output now includes the
actual error from sub-tool commands instead of generic
hardcoded messages.

## Before

### uf init

```
Sub-tool initialization:
  ✗ .specify/ failed (specify init failed)
  ✗ dewey index failed (dewey index failed)
  ✓ .uf/replicator/ initialized
```

The user must re-run `specify init .` and `dewey index`
manually to discover the actual errors.

### uf setup

```
  ✗ gaze             failed (brew install failed)
                     Error: brew install gaze: exit status 1
```

The `Error:` line only shows "exit status 1" -- not what
brew actually printed.

## After

### uf init

```
Sub-tool initialization:
  ✗ .specify/ failed (specify init: exec: "specify": executable file not found in $PATH)
  ✗ dewey index failed (dewey index: workspace not initialized)
                        Output: Error: no vault found at current directory.
                                Run 'dewey init --vault .' to create one.
  ✓ .uf/replicator/ initialized
```

The failure message includes the sub-tool's actual error.
When command output is available, it is shown on
subsequent indented lines.

### uf setup

```
  ✗ gaze             failed (brew install failed)
                     Error: brew install gaze: exit status 1
                     Output: Error: No available formula with the name "gaze".
                             ==> Searching for a previously deleted formula...
```

The new `Output:` line shows what the package manager
actually reported.

### Long output (> 20 lines)

```
  ✗ dewey            failed (go install failed)
                     Error: go install github.com/unbound-force/dewey@latest: exit status 1
                     Output: ... (47 lines omitted)
                             cannot find module providing package ...
                             ... (last 10 lines of actual output)
```

Output exceeding 20 lines is truncated to the last 10
lines with a count of omitted lines.

## Successful runs

No change. Successful steps continue to show the same
single-line summary:

```
  ✓ .specify/ initialized
  ✓ dewey index completed
```

No additional output is produced for passing steps
(FR-007).

## Affected Commands

| Command    | Package           | Change |
|------------|-------------------|--------|
| `uf init`  | internal/scaffold | Error output in sub-tool failure messages |
| `uf setup` | internal/setup    | Error output in tool installation failure messages |
