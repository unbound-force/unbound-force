## Context

`uf init` scaffolds `.opencode/` files into a target
repository. When command files were renamed from bare
names to `uf.`-prefixed names (PR #304, commit `e3b4bb7`),
a `cleanupRenamedCommands()` function was added to remove
old-name files. The function exists at
`internal/scaffold/scaffold.go:327-339` and is called
unconditionally at line 249.

Current problems:

1. `cleanupRenamedCommands()` silently swallows all errors
   from `os.Stat` and `os.Remove` -- if removal fails
   (permissions, race condition), no diagnostic is emitted.
2. The function has zero test coverage, so the failure
   mode was never caught.
3. Agent files referencing old command names (e.g.,
   `/review-council` instead of `/uf.review-council`) are
   never updated because `isToolOwned()` classifies
   `opencode/agents/` as user-owned.

The `Result` struct already has a `Migrated []string`
field (line 54) and the returned paths are appended at
line 250. The `printSummary` function renders file
categories including migrated files. The plumbing for
reporting already exists -- the function just needs
error visibility.

## Goals / Non-Goals

### Goals
- Make `cleanupRenamedCommands()` failures visible via
  structured logging so users can diagnose why old files
  persist
- Add test coverage for the cleanup happy path and error
  paths
- Warn users about stale command references in agent
  files, following the `warnLegacyReviewerFiles()`
  precedent at line 260

### Non-Goals
- Auto-rewriting agent files: agents are user-owned by
  design (`isToolOwned()` returns false for
  `opencode/agents/`). Silently modifying user content
  violates the ownership model. A warning is sufficient.
- Changing `isToolOwned()` classification for agents:
  this would break the customization model for all users,
  not just those migrating from old command names.
- Adding a `--migrate-agents` flag: out of scope for a
  bug fix. Could be considered in a future enhancement.

## Decisions

### D1: Log errors, do not fail the scaffold run

`cleanupRenamedCommands()` currently returns
`[]string` (removed paths). The function will be
modified to also log warnings for files that could
not be removed. It will NOT return an error or cause
`Run()` to fail -- cleanup is best-effort. The scaffold
run should complete and report what it could not clean.
The function is idempotent — re-running `uf init` after
a partial cleanup will attempt to remove any remaining
old files.

Rationale: Observable Quality (Principle III) requires
failures to be surfaced. But the scaffold run should
not abort because of a cleanup failure -- the new files
were already created successfully.

### D2: Accept `io.Writer` for warning output

`cleanupRenamedCommands()` currently takes only
`targetDir string`. Add `io.Writer` as a parameter
(matching the pattern used by `warnLegacyReviewerFiles`)
so warnings are testable via buffer injection. The
caller at line 249 passes `opts.Stdout`.

### D3: New `warnStaleCommandRefs()` function

Create a new function that:
1. Walks `opencode/agents/*.md` files in the target dir
2. Reads each file and scans for old command name
   patterns (the keys of `renamedCommands`, stripped to
   base names without `.md`, prefixed with `/`)
3. For each match, records the file and stale reference
4. Prints a warning block listing all stale references
   and their correct replacements

This follows the exact precedent of
`warnLegacyReviewerFiles()` (line 260-274): scan,
warn, do NOT modify.

### D4: Test strategy

Tests for `cleanupRenamedCommands()`:
- Happy path: create old-name files in `t.TempDir()`,
  run cleanup, verify files removed and paths returned
- No-op path: no old files exist, verify empty return
- Partial failure: make one file read-only, verify
  remaining files are still cleaned and warning is
  logged
- Verify the returned paths use the mapped output paths
  (`.opencode/commands/...`), not the internal asset
  paths (`opencode/commands/...`)

Tests for `warnStaleCommandRefs()`:
- Agent file with stale ref: verify warning output
  contains file name and old/new command mapping
- Agent file with no stale refs: verify no warning
- No agent files: verify no warning and no error

All tests use `t.TempDir()` and `bytes.Buffer` for
isolation (Principle IV: Testability).

## Risks / Trade-offs

### Risk: Silent permission errors on cleanup

If a file is owned by root or read-only, `os.Remove`
fails. The current code swallows this silently. The fix
adds logging but does not retry or escalate. Users on
systems with restrictive permissions will need to remove
files manually after seeing the warning.

Accepted: logging is sufficient. Automatic privilege
escalation would violate Security by Default
(Principle V).

### Risk: Stale reference patterns may have false positives

Scanning agent files for `/review-council` could match
documentation text or comments, not just actual command
references. The warning may occasionally flag non-command
text.

Accepted: false positives in a warning are low-cost.
The warning lists affected files and the old/new command
references so users can judge. No files are modified.

### Trade-off: Warning vs. auto-fix for agents

Auto-fixing agent files would provide a better migration
experience but would violate the tool-owned/user-owned
boundary. Users who customized their agent files would
have their changes silently modified. The warning
approach is conservative but safe.
