<!-- originating_issue: 903 is recorded in .openspec.yaml; /uf.finale
     emits `Closes #903` in the PR body. Issue #902 is in scope of
     this same change and the PR MUST close it too -- the PR body
     carries `Closes #902` alongside the finale-emitted `Closes
     #903` (see the bookkeeping task in tasks.md). The repo
     documentation issue filed per the AGENTS.md documentation gate
     is filed during implementation; its number is added to the PR
     body at that time. -->

## Why

Two scaffold defects, one surface. Both land on the
`ensureGitignore` / ensure-side-effects seam of the scaffold
engine, so they ship as one cohesive change.

**#903 -- override stubs.** Both per-developer override files --
`.uf/model-matrix.override.yaml` and
`.uf/review-matrix.override.yaml` -- are valid no-ops when
absent: model-matrix profiles resolve to the host model and
review dispatch falls back to tier defaults. That means a fresh
`uf init` leaves every developer on host-model defaults until
they hand-create an override file from scratch, copying the
whitelist rules out of the docs. The tracked matrices already
point developers at the override files in their header comments;
the scaffold should close that loop by seeding them. The
gitignore side is already done -- both entries ship in the
scaffold's managed `gitignoreBlock` and the repo root
`.gitignore` -- so this change adds only regression pins there.

**#902 -- gitignore block drift.** `uf init` appends a managed
block to the target's `.gitignore` via `ensureGitignore()`,
idempotent on the marker `# Unbound Force -- managed by uf
init`. The block content, `gitignoreBlock`
(scaffold.go:1399-1421), has drifted from the meta repo's own
`.gitignore`: six patterns added to the repo over the past
months never made it into the scaffold, so fresh `uf init`
targets do not ignore them. Consequences: fresh targets show
untracked runtime noise, risk committing per-developer overrides
and runtime data, and the scaffolded AGENTS.md sibling section
(scaffold.go:1744) states that clones and caches live under
`.uf/sibling-clones/` / `.uf/sibling-cache/` as "gitignored,
in-workspace paths" while `gitignoreBlock` does not ignore them
-- an explicit contradiction in generated output. Secondary
finding: the meta repo tracks `.opencode/.gitignore`
(`node_modules`, `bun.lock`) but nothing deploys it, while
`uf init` runs `npm ci` in `.opencode/`
(plugin_activation.go:69), so fresh targets get an untracked
`.opencode/node_modules/` unless their pre-existing root ignore
happens to cover it.

**Contract reversal.** Seeding the override files reverses a
recorded stance: SC-FR-010 in
`openspec/changes/model-matrix-scaffold-deployment/` says the
override "MUST remain preserved and never scaffolded beyond the
gitignore entry", and D10 in
`openspec/changes/model-matrix-delegation/design.md` says the
same. The never-scaffolded stance protected per-developer files
from team-init churn. Create-only-if-absent seeding preserves
that protection entirely -- the file is only ever written when
it does not exist -- while removing the onboarding friction.
This change carries a MODIFIED delta on SC-FR-010 and records
the D10 amendment in its design.

Done means (issues #903 and #902): a fresh `uf init` in an
empty directory creates both override stubs with documented
whitelist comments and a managed gitignore block covering
every repo pattern in the managed region (`.gaze/` stays
out -- gaze is not deployed by `uf init`); pre-existing
override files survive plain re-runs
and `--force` byte-for-byte; `uf init --overrides` creates
exactly the two stubs plus the gitignore block and nothing
else, and errors when combined with `--divisor`, `--force`,
`--check`, or `--stealth`; stale managed blocks in existing
repos upgrade in
place on the next init; `.opencode/.gitignore` deploys as a
sub-tool; the repo's own `.gitignore` is deduplicated and its
managed block pinned byte-for-byte to `gitignoreBlock` by a
drift test; docs and CHANGELOG are updated; `make check` passes
end to end.

## What Changes

1. **Override stub seeding (full mode)** -- `uf init` (full
   mode) seeds both per-developer override stubs,
   create-only-if-absent, as comment-only Go-constant templates
   (like `gitignoreBlock`, deliberately NOT embedded assets:
   avoids the 8-piece asset-registration checklist, and the
   files are gitignored per-developer data with no canonical
   tracked source for drift comparison). Each stub documents
   its file's whitelisted paths with commented examples
   (model-matrix: `profiles.*.model`, `profiles.*.variant`,
   `classifier.llm.*`; review-matrix:
   `profiles.<tier>.model`/`variant`). Comment-only content
   parses to YAML null, which both loaders already treat as a
   valid no-op (the model-matrix empty-override scenario and
   RD-FR-020's comments-only scenario) -- zero loader changes.

2. **Override preservation** -- existing override files are
   preserved byte-for-byte on plain re-runs AND under `--force`
   (per-developer data; `--force` resets tracked canonical
   files only). Seeding is skipped under `--stealth` (consistent
   with the other ensure-side-effects) and in `--divisor` mode.
   New `ensureOverrideStubs` modeled on `ensureGitignore`;
   `Options.OverridesOnly` gates the new mode.

3. **`uf init --overrides` mode** -- creates ONLY the two
   override stubs plus the managed `.gitignore` block
   (idempotent); skips the embedded-asset walk, sub-tools,
   AGENTS.md sections, and migrations entirely. Mutually
   exclusive with `--divisor`, `--force`, `--check`, and
   `--stealth` (conflict validation in
   `cmd/unbound-force/main.go` plus help text).

4. **Contract reversal** -- a MODIFIED delta on SC-FR-010 moves
   the override file from "never scaffolded beyond the gitignore
   entry" to "seeded as a create-only-if-absent comment-only
   stub, never modified afterwards (not even by `--force`)";
   the D10 amendment and its rationale are recorded in this
   change's design.

5. **gitignoreBlock sync** -- the six missing patterns are
   added, grouped semantically with the existing sections:
   `.uf/config.yaml`, `.uf/dewey/config.yaml`,
   `.uf/sibling-clones/`, `.uf/sibling-cache/`,
   `.uf/feedback/`, `.uf/progress/`. `.gaze/` stays out --
   gaze is a separate hero not deployed by `uf init`.

6. **Managed-block replacement upgrade** -- `ensureGitignore()`
   and stealth mode's derived `.git/info/exclude` writer gain
   upgrade semantics: when the marker is present but the block
   differs, the marker -> `.unbound-force/` region is rewritten
   in place (deterministic, no duplicate blocks, trailing
   content preserved); a malformed block (marker without
   terminator) is skipped with detail; each run reports
   created / upgraded / unchanged per file.

7. **`.opencode/.gitignore` sub-tool** -- scaffolded from Go
   code (Go's `//go:embed` excludes dotfiles from directory
   matches, so the assets tree cannot ship it) as a new sub-tool
   independent of the plugin-activation result, idempotent,
   preserving unknown user lines. Content: `node_modules`,
   `bun.lock`, `.vitest/`, `coverage/`. The canonical
   `.opencode/.gitignore` and the
   `TestPluginPackageInputs_GitignoreTracksCanonicalInputs`
   want-string (scaffold_test.go:1616) are updated to match.

8. **Drift test (recurrence prevention)** -- a scaffold test
   asserts the repo's own managed block (marker ->
   `.unbound-force/` region) equals `gitignoreBlock`
   byte-for-byte, reusing the same extraction helper the
   replacement logic uses, so block drift can never recur
   silently.

9. **Repo `.gitignore` sync** -- deduplicate
   `.uf/dewey/config.yaml` (duplicated at lines 40 and 47),
   pluralize the per-developer override comment, and drop the
   now-redundant root `.opencode/.vitest/` and `coverage/`
   lines (the scaffolded `.opencode/.gitignore` covers them).

10. **Docs and bookkeeping** -- `docs/cli-reference.md` `uf
    init` flag table gains `--overrides` with an example;
    `docs/configuration.md` updates the four "never seeded"
    passages (lines 275, 286-288, 471, 487) to the new
    create-only-if-absent contract; `CHANGELOG.md` gains an
    entry.

Out of scope: any loader change (model-matrix.ts,
review-dispatch -- comment-only files are already valid
no-ops); embedded-asset registration for the stubs (they are
Go constants); `.gaze/` (separate hero); the
`stealthExcludePaths` canonical set (unchanged); and the
review-matrix / reviewer-capabilities ownership semantics
(#884/#886 precedent, unchanged).

**Sequencing**: this change's MODIFIED delta on SC-FR-010
builds on the requirement the `model-matrix-scaffold-
deployment` change defined; that change is complete (19/19
tasks) but not yet archived, so it MUST archive before this
change does. Its PR landing first is the natural order and
requires no coordination beyond normal queueing.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `scaffold`: `uf init` full mode seeds both per-developer
  override stubs create-only-if-absent (comment-only, YAML-null
  no-op), preserved byte-for-byte across re-runs and `--force`,
  skipped under `--stealth` and `--divisor`; a new
  `--overrides` mode creates only the stubs plus the managed
  gitignore block; `gitignoreBlock` gains the six missing repo
  patterns; managed blocks (gitignore and stealth exclude)
  upgrade in place when stale, reporting created / upgraded /
  unchanged; `.opencode/.gitignore` deploys as an idempotent
  sub-tool; a drift test pins the repo's own managed block to
  `gitignoreBlock`. SC-FR-010's override clause is MODIFIED to
  the seeded-once contract.

### Removed Capabilities

- None.

## Impact

- **Modified**: `internal/scaffold/scaffold.go`
  (`gitignoreBlock` content, `ensureGitignore` upgrade
  semantics, new `ensureOverrideStubs` and `.opencode/.gitignore`
  writer, `Options.OverridesOnly`, `Run` wiring);
  `internal/scaffold/scaffold_stealth.go` (exclude-block
  replacement semantics); `internal/scaffold/scaffold_test.go`
  (new seeding / preservation / mode-isolation / upgrade /
  drift tests, representative-patterns extension,
  want-string update); `cmd/unbound-force/main.go`
  (`--overrides` flag, conflict validation, help text);
  `.gitignore` (dedupe, comment pluralization, redundant-line
  removal); `.opencode/.gitignore` (canonical content);
  `docs/cli-reference.md`; `docs/configuration.md`;
  `CHANGELOG.md`.
- **Unchanged**: the matrix and review-matrix loaders and their
  override merge semantics; the embedded-asset walk and its
  file-count arithmetic (the stubs are Go constants, not
  assets); every existing ownership classification other than
  the override stub clause; the stealth canonical path set.
- **Behavior**: fresh targets receive both override stubs and
  complete ignore coverage out of the box; existing repos with
  stale managed blocks upgrade in place on the next `uf init`;
  `--force` never touches per-developer override files.
- **Security**: the stubs are comment-only documentation with
  no credential-shaped content; no secrets, keys, or model
  identities are deployed.
- **Gates**: no coverage threshold, CRAP score, CI flag, or
  pinned dependency changes; the `.opencode/` suite must hold
  its governance-controlled 94% statement coverage floor, and
  `make check` (lint, test, build, coverage gate, plugin tests)
  must pass end to end.
- **Documentation gate**: user-facing change -- BOTH gates
  fire before PR merge (final phase task): the repo
  documentation issue AND the constitution v1.4.0
  Cross-Repo Documentation tracking issue in
  unbound-force/website.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0,
amended 2026-10-14; six active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The stubs are self-describing artifacts: each documents its own
whitelist and merge semantics in comments, so a developer (or
agent) learns the override contract from the file itself
without session reconstruction. The gitignore drift test turns
an implicit cross-file assumption (repo block == scaffold block)
into a checked invariant. No runtime coupling between heroes is
introduced; `.gaze/` stays out of the scaffolded block because
gaze is not deployed by `uf init`.

### II. Composability First

**Assessment**: PASS

Every seeded piece remains independently removable: the
override stubs are valid no-ops in both loaders (an existing
contract), `uf init --overrides` works without any prior
scaffold, and the `.opencode/.gitignore` sub-tool runs
independently of plugin activation. No new mandatory
dependencies; the change adds no coupling between the scaffold
and the loaders.

### III. Observable Quality

**Assessment**: PASS

Each ensure-side-effect reports a machine-readable disposition
(created / upgraded / unchanged / skipped with detail) through
the existing `subToolResult` summary, so what a run did to
`.gitignore`, the stealth exclude, and each override stub is
inspectable without diffing trees. The stale-block upgrade path
is deterministic and reported, never silent.

### IV. Testability

**Assessment**: PASS

Every requirement is verifiable in isolation with the existing
scaffold test seams (`Options.ReadFile`/`WriteFile` injection,
`t.TempDir()` targets, real `Run()` invocations): exact-content
creation, byte-for-byte preservation including `--force`,
mode isolation, conflict validation, upgrade rewriting, and
the drift test all run without external services.

### V. Security by Default

**Assessment**: PASS

The stubs deploy no secrets -- comment-only documentation with
no credential-shaped content -- and the seeded files are
gitignored per-developer data, so provider-specific model
strings never enter version control through the scaffold.
Ignore-coverage expansion (six patterns) reduces the risk of
committing per-developer overrides and runtime data.

### VI. Deterministic Execution

**Assessment**: PASS

Seeding, preservation, and block replacement are deterministic
file operations with explicit dispositions -- identical inputs
produce identical trees. The managed-region extraction helper
is shared between the replacement logic and the drift test, so
the behavior cannot drift from its own verification.
