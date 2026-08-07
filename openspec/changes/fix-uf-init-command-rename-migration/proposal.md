## Why

`uf init` does not reliably remove old-name command files
when migrating to the `uf.` namespace prefix. Running
`uf init` on a repo with pre-rename command files (e.g.,
`address-feedback.md`) creates the new `uf.*` files
alongside the old ones, leaving duplicates. The
`cleanupRenamedCommands()` function exists but silently
swallows errors, making failures invisible. Additionally,
agent files that reference old command names are never
updated because `isToolOwned()` classifies agents as
user-owned.

This was reported as issue #419 with reproduction evidence
from the eval-infra repository showing that old-name files
persist after re-running `uf init` with a post-rename
binary.

## What Changes

### Fix 1: Make `cleanupRenamedCommands()` observable

The function currently swallows all errors from `os.Stat`
and `os.Remove`. When cleanup fails, the user sees no
indication that old files remain. Add warning output
using `fmt.Fprintf` to the injected `io.Writer` so
failures are reported in the scaffold summary.

### Fix 2: Add test coverage for cleanup path

`cleanupRenamedCommands()` has zero test coverage. Add
tests that:
- Create a temp directory with old-name command files
- Run `cleanupRenamedCommands()`
- Verify old files are removed
- Verify the returned list matches removed files
- Test the error path (read-only files, missing files)

### Fix 3: Warn about stale command references in agents

Since agent files are user-owned by design (users
customize them), `uf init` cannot silently overwrite
them. Instead, add a warning pass that scans agent files
for references to old command names and prints a
diagnostic listing which files contain stale references
and what the new names are. This follows the precedent
set by `warnLegacyReviewerFiles()`.

## Capabilities

### New Capabilities
- `stale-reference-warning`: `uf init` warns when agent
  files contain references to old (pre-namespace-prefix)
  command names, listing each file and the stale
  references found

### Modified Capabilities
- `cleanupRenamedCommands`: errors are now logged instead
  of silently swallowed; removed files appear in the
  scaffold summary output

### Removed Capabilities
- None

## Impact

- `internal/scaffold/scaffold.go`: modify
  `cleanupRenamedCommands()` to log errors; add new
  `warnStaleCommandRefs()` function
- `internal/scaffold/scaffold_test.go`: add test cases
  for `cleanupRenamedCommands()` and
  `warnStaleCommandRefs()`
- No changes to public API, CLI flags, or artifact
  schemas
- No new dependencies required

### Documentation Impact
- `CHANGELOG.md`: entry needed under bug fixes for
  cleanup error visibility and stale reference warnings
- `AGENTS.md`: no changes needed (no structural changes)
- Website docs issue: exempt — the warning output is
  self-explanatory and does not change CLI flags or
  commands

## Constitution Alignment

Assessed against the Unbound Force org constitution v1.2.0.

### I. Autonomous Collaboration

**Assessment**: N/A

This change is internal to the scaffold engine. It does
not affect artifact-based communication between heroes
or alter any inter-hero exchange formats.

### II. Composability First

**Assessment**: PASS

The fix maintains standalone functionality of `uf init`.
No new dependencies are introduced. The warning for
stale agent references is informational only and does
not require any other hero to be present.

### III. Observable Quality

**Assessment**: PASS

This change directly improves observability: silent
error swallowing is replaced with structured logging,
and a new diagnostic warning surfaces stale references
that were previously invisible. The scaffold summary
output gains additional machine-parseable information
about migrated files.

### IV. Testability

**Assessment**: PASS

The primary deliverable includes test coverage for a
previously untested function. Tests verify observable
side effects (file removal, returned path lists, warning
output) rather than implementation details. All tests
use `t.TempDir()` for filesystem isolation.

### V. Security by Default

**Assessment**: N/A

This change does not introduce dependencies, handle
external inputs beyond the existing file paths, or
modify privilege boundaries. The file operations
(`os.Stat`, `os.Remove`) already operate within the
target directory scope.
