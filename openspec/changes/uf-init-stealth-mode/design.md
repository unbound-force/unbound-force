## Context

`uf init` scaffolds `.opencode/`, `openspec/`, `.specify/`, `.uf/`, and
`opencode.json`, then unconditionally mutates two tracked files:

- `ensureGitignore` (`internal/scaffold/scaffold.go:1130`) appends the
  `gitignoreBlock` (marker `# Unbound Force — managed by uf init`) to
  `.gitignore`, creating it if absent.
- `ensureAGENTSmdPackSection` (`scaffold.go:1336`) appends a
  "Convention Packs" section to `AGENTS.md`.

`Run` calls both unconditionally (lines 237 and 240), independent of
`DivisorOnly`. For a local-only user authoring specs and PRs against an
upstream project that does not adopt unbound-force, every scaffolded
file appears in `git status` and risks being committed or shipped in a
PR (issue #638, discussion #523).

## Goals / Non-Goals

### Goals

- Provide a `--stealth` mode that leaves the working tree git-clean
  (`git status --porcelain` empty) and all tracked files unmodified,
  while pipeline commands still resolve scaffolded files normally.
- Keep normal (non-stealth) init behavior byte-identical.
- Make the git-clean invariant observable (`--check` and a post-init
  summary, both with machine-parseable output) and failure-atomic.

### Non-Goals

- Reconcile with #198 (XDG symlink farm for tool-owned files). That is
  a larger out-of-tree storage redesign; stealth mode must not depend on
  it. It is noted as a future consolidation point.
- Introduce any new external dependency.
- Change the behavior of `--divisor`, `--force`, `--lang`, or
  `--dry-run`.
- Provide an `--unstealth` teardown command (deferred as a non-goal).
  Manual recovery — removing the marker-delimited `.git/info/exclude`
  block — is documented in the help text.

## Decisions

### D1 — Add a `Stealth` field to `scaffold.Options` and a `--stealth` flag

Slots into the existing struct-based flag pattern (`Force`,
`DivisorOnly`, `DryRun`, `Lang`). The CLI wires it through
`initParams` → `scaffold.Options` (mirroring `--divisor`). Default is
`false`, so non-stealth behavior is unchanged. `Stealth` is distinct
from `DryRun`: `DryRun` suppresses `opencode.json` writes only, while
stealth writes everything but keeps it git-invisible.

A separate `--check` flag (see D5) is stealth-only and distinct from
the pre-existing `--dry-run`. `--check` persists nothing at all,
whereas `--dry-run` persists scaffolded files while suppressing
`opencode.json`. The `--check` flag maps to a `Check bool` field on
`Options`, mirroring the `Stealth` field.

### D2 — Use `.git/info/exclude` as the exclusion mechanism

Stealth mode appends the scaffolded top-level paths (`.opencode/`,
`.specify/`, `openspec/`, `.uf/`, `opencode.json`) to
`.git/info/exclude` instead of `.gitignore`. Rationale:

- `.git/info/exclude` is a local-only git facility — never committed,
  never touched by `uf init`'s `.gitignore` writes — satisfying
  "tracked files unmodified."
- Files remain physically in the repo at their conventional paths, so
  `isToolOwned` drift-detection, `migrateCommandDir`, and every pipeline
  command's path resolution continue to work with **no indirection**.
  This avoids the reliability ripple (path-resolution/drift-detection
  must follow symlinks/overlay) that an out-of-tree farm (#198) would
  introduce.

The exclusion path MUST be resolved via `git rev-parse --git-path
info/exclude` (not a hardcoded `.git/info/exclude`), so linked
worktrees and submodules — where `.git` is a file, not a directory —
write to the correct location.

The exclusion list MUST be complete for all output that `uf init` and
its sub-tools create. Stealth mode records any newly-created untracked
paths during init and appends them to the marker-delimited block, so
files scaffolded outside the five canonical paths are also hidden.
Appended paths are emitted as literal patterns (leading `#`/`!`
escaped, no trailing whitespace) so a path beginning with a gitignore
metacharacter is not interpreted as a directive.

Trade-off: `.git/info/exclude` is non-portable (lost on re-clone) and
invisible to tools that do not honor it. This is acceptable because
stealth is *by definition* local-only and not meant to be recreated on
clone. This consequence is documented in the command help text and is
asserted by a test.

### D3 — Protect tracked files against sub-tool edits via snapshot/restore

`uf init` delegates to sub-tools (`initSubTools`, `initDewey`) that may
also edit `.gitignore` (issue #638 body, point 2) and potentially other
tracked files. Stealth mode MUST cover more than `uf init`'s own
writes. Rather than enumerate which sub-tools touch which tracked
files, stealth mode snapshots the content bytes **and mode/permission
bits** of all tracked files (permission bits obtained via a per-file
`os.Lstat` — not `os.Stat`, which follows symlinks — so full permission
bits and symlink mode `120000` are preserved; the index executable bit
is additionally available from `git ls-files -s -z` stage mode, before
sub-tool delegation) and restores any that differ afterward, preserving
the original permission bits on restore. This directly satisfies the
"no modified tracked file" half of the clean-tree invariant and does
not require a per-tool audit. The snapshot cost is proportional to the
number of tracked files; stealth init is a rare, non-hot-path
operation, so this is acceptable. Tracked symlinks (mode `120000`) are
restored as symlinks (re-reading the link target from the snapshot),
never by writing their content bytes as a regular file. A future
optimization may restore committed files from the git object store
(`git restore`) rather than holding bytes in memory.

### D4 — Failure atomicity

The snapshot/restore from D3 doubles as rollback for tracked files. In
addition, stealth mode orders its writes so the exclusion entry is
written before scaffolding, and on any failure it removes any
scaffolded files it has written, removes only the marker-delimited
exclusion block appended by the current invocation (preserving any
pre-existing block), restores the index to its full init-start
snapshot (path → blob + mode), un-staging paths newly staged during
this init (currently staged minus staged at init start, never a
blanket `git reset`), restores tracked files to their pre-init bytes
and mode/permission bits, and reports a clear error. Stealth mode MUST
NOT leave the tree half-applied in a way that could cause unbound-force
artifacts to be committed unintentionally. After a failed stealth
init, `git status --porcelain` MUST be empty. If a rollback step itself
fails, stealth mode MUST report a clear error naming the residual dirty
paths rather than silently asserting cleanliness.

### D5 — Observable cleanliness

Stealth mode adds a post-init summary reporting `git status --porcelain`
emptiness (or the specific files that remain visible) and a `--check`
verification path that reports the expected tree state without
persisting anything. `--check` MUST be combined with `--stealth` in
the same invocation. Both the `--check` output and the post-init
summary emit a machine-parseable (JSON) result per Constitution III.
`--check` is also re-runnable after init (still combined with
`--stealth`): it detects a missing or stale `.git/info/exclude` block
and reports which scaffolded paths have become visible, so a wiped
exclusion is caught rather than silently re-exposing `.uf/` (which may
hold secrets). `--check` exits non-zero when the tree is not clean or
the exclusion block is missing/stale, and exits 0 only when the
invariant holds, so it is usable as an automation gate. Normal (non-
stealth) `uf init` warns when it detects a stale marker-delimited
stealth exclusion block, since lingering exclusions would otherwise
silently hide scaffolded files from `git add`.

### D6 — Stealth requires a git repository

Stealth mode depends on `.git/info/exclude`, which requires being
inside a git repository. Stealth init MUST detect "not inside a git
repo" via `git rev-parse --is-inside-work-tree` (not a literal `.git`
directory check, since linked worktrees and submodules have `.git` as
a file), report a clear error, and write no files before any
scaffolding begins. The probe MUST distinguish "git not installed / not
on PATH" from "not inside a work tree" and report a specific, actionable
error for each. This is distinct from non-stealth `uf init`, which does
not require git.

## Test Strategy

Coverage strategy for the new stealth code path, per Constitution IV:

- **Unit tests** (no real git binary, `t.TempDir()` + injected
  `ReadFile`/`WriteFile`/`ExecCmd` fakes):
  - `writeGitInfoExclude` — idempotent append via marker, path
    resolution through `git rev-parse --git-path info/exclude`,
    unwritable-target error handling.
  - `snapshotTrackedFiles` / `restoreTrackedFiles` — capture and restore
    of tracked-file bytes AND mode/permission bits (including a
    non-default-mode file), a tracked symlink restored as a symlink,
    plus a sub-tool `.gitignore` edit (only `ExecCmd` is faked so the
    restore is asserted against real on-disk bytes and modes).
  - `Stealth` branch wiring — skips `ensureGitignore` /
    `ensureAGENTSmdPackSection`, calls `writeGitInfoExclude`, restores on
    mid-`Run` failure.
  - `--check` — reports cleanliness and persists nothing.
- **Integration tests** (real `git` binary in `t.TempDir()`): assert
  `git status --porcelain` is empty after stealth init; assert a
  non-git directory errors and writes nothing; assert already-tracked
  scaffolded paths are detected and refused.
- **e2e tier**: not applicable to this scaffold-engine path (no
  long-running service or multi-process flow).
- **Coverage target**: ≥ 90% statement coverage on
  `writeGitInfoExclude`, `snapshotTrackedFiles`, `restoreTrackedFiles`,
  and the stealth branches of `Run`, enforced by the existing coverage
  ratchet in `ci_local.yml`. A coverage regression blocks the build.

The existing normal-mode regression test (task 3.4) asserts non-stealth
behavior is byte-identical to the current implementation, guarding
against accidental change to the default path.

## Risks / Trade-offs

- **`.git/info/exclude` non-portability**: accepted (local-only feature);
  documented in help text, and asserted by a test, so users do not
  expect re-clone to preserve it.
- **Sub-tool edits beyond our control**: mitigated by D3's snapshot of
  all tracked files; residual risk if a sub-tool writes outside the
  snapshot window is covered by the post-init summary (D5) flagging any
  dirty tree.
- **Already-tracked scaffolded paths**: `.git/info/exclude` only hides
  untracked files. Stealth mode detects scaffolded paths that are
  already tracked and refuses with a clear error rather than silently
  breaking the git-clean invariant.
- **Flag semantics confusion with `--dry-run`**: mitigated by distinct
  naming (`--check`) and help text documenting the difference.
- **Interaction matrix** (`--stealth` × `--force` × `--divisor` ×
  `--lang` × `--dry-run` × `--check`): `--stealth --force` is valid
  (overwrite + hide); `--stealth --divisor` composes (subset deployment
  + hidden). The full matrix is specified in the spec's "Flag
  interactions" requirement, not enforced by new logic, since stealth
  only changes *where* the exclusion is written.
