<!-- capability: scaffold
     change: init-override-stubs-gitignore-sync
     originating_issue: 903 -->

## Context

Both issues land on the same seam of the scaffold engine: the
ensure-side-effects that run after the embedded-asset walk --
`ensureGitignore` (scaffold.go:1432), the AGENTS.md sections,
and the stealth-mode `.git/info/exclude` writer
(`writeGitInfoExclude`, scaffold_stealth.go:163).

**#903.** The per-developer override files
(`.uf/model-matrix.override.yaml`,
`.uf/review-matrix.override.yaml`) are valid no-ops when
absent, so a fresh `uf init` leaves developers on host-model
defaults until they hand-create an override from the docs. The
tracked matrices already point at the override files in their
header comments. The gitignore entries already ship in
`gitignoreBlock`; only seeding is missing.

**#902.** `gitignoreBlock` (scaffold.go:1399-1421) has drifted
from the repo's own `.gitignore`: six patterns
(`.uf/config.yaml`, `.uf/dewey/config.yaml`,
`.uf/sibling-clones/`, `.uf/sibling-cache/`, `.uf/feedback/`,
`.uf/progress/`) never made it into the scaffold. The
scaffolded AGENTS.md sibling section (scaffold.go:1744)
explicitly contradicts the block. And `.opencode/.gitignore`
is tracked in this repo but never deployed, while `uf init`
runs a lockless `npm install` in `.opencode/`
(plugin_activation.go:69), so
fresh targets get an untracked `.opencode/node_modules/`.

Two prior contracts record a never-scaffolded stance for the
model-matrix override: SC-FR-010 in
`openspec/changes/model-matrix-scaffold-deployment/` and
decision D10 in
`openspec/changes/model-matrix-delegation/design.md`. This
change reverses that stance to seeded-once (the MODIFIED delta
in specs/scaffold/spec.md) and records the amendment here.

Constitution alignment is assessed in proposal.md (all six
principles PASS); the decisions below reference the principles
they touch.

## Goals / Non-Goals

### Goals

- Seed both override stubs create-only-if-absent in full-mode
  `uf init`, with comment-only content that both loaders
  already accept as a no-op.
- Preserve existing override files byte-for-byte across plain
  re-runs and `--force`.
- Add `uf init --overrides` for seeding only the stubs plus
  the managed gitignore block, with fail-closed flag conflicts.
- Sync `gitignoreBlock` with the six missing repo patterns
  (`.gaze/` excluded) and pin repo-to-scaffold equality with a
  drift test.
- Upgrade stale managed blocks (gitignore and stealth exclude)
  in place, deterministically, reporting created / upgraded /
  unchanged / skipped per run.
- Deploy `.opencode/.gitignore` from Go code as an idempotent
  sub-tool independent of plugin activation.
- Amend SC-FR-010 and record the D10 reversal with rationale.
- Update `docs/cli-reference.md`, `docs/configuration.md`, and
  `CHANGELOG.md`; keep `make check` green.

### Non-Goals

- Any loader change: `model-matrix.ts` and the review-dispatch
  override merge are untouched (comment-only files are already
  valid no-ops per the model-matrix empty-override scenario
  and RD-FR-020's comments-only scenario).
- Embedded-asset registration for the stubs or the
  `.opencode/.gitignore` file: both are Go constants, so the
  embedded-asset walk and its file-count arithmetic are
  unchanged.
- `.gaze/` in the scaffolded block (gaze is a separate hero
  not deployed by `uf init`).
- Changes to `stealthExcludePaths` (the canonical top-level
  path set is current).
- Changes to `--force` semantics for tracked files, to the
  review-matrix / reviewer-capabilities ownership
  classifications (#884/#886 precedent), or to the repo root
  `.gitignore` content outside the managed region except the
  two redundant noise lines noted in D9.

## Decisions

### D1: Stubs are Go string constants, not embedded assets

`modelMatrixOverrideStub` and `reviewMatrixOverrideStub` are
Go string constants in `internal/scaffold/scaffold.go`, the
`gitignoreBlock` precedent. Embedded assets were rejected
because the files are gitignored per-developer data with no
canonical tracked source for drift comparison, and embedding
would drag in the 8-piece asset-registration checklist
(embedded asset list, drift table, `expectedAssetPaths`,
DivisorOnly filtering, marker insertion) for content that
never round-trips. Consequence (pinned by tests): the
embedded-asset file-count arithmetic does not change.

### D2: Comment-only stub content (YAML null)

Each stub documents its whitelist with commented-out examples
(model-matrix: `profiles.*.model`, `profiles.*.variant`,
`classifier.llm.*`; review-matrix:
`profiles.<tier>.model`/`variant`) and nothing else. A
comment-only file parses to YAML null; both loaders already
treat null as a valid no-op, so seeding requires zero loader
changes. The stubs are self-describing artifacts (Constitution
I): a developer learns the override contract from the file
itself.

### D3: Create-only-if-absent, preserved even under --force

`ensureOverrideStubs` writes each stub only when the target
file does not exist; any existing file is reported skipped and
never read-modified. `--force` does not bypass this:
per-developer data is never reset by the scaffold, which
resets tracked canonical files only. This is the load-bearing
reversal decision -- see "Amendment of prior decisions" below.

### D4: Skip surfaces for stub seeding

Seeding is skipped under `--stealth`: stealth is a local-only
trial whose rollback sweeps `.uf/`, so per-developer state
must not be seeded into it. It is skipped in `--divisor`
mode: a minimal review-council subset deployment into
possibly-foreign repositories does not write per-developer
files. In `Run()`, `ensureOverrideStubs` is called in the
existing non-stealth ensure-branch, after `ensureGitignore`,
gated on `!opts.DivisorOnly`.

### D5: Overrides-only mode as an early-exit path

`Options.OverridesOnly` gates a path in `Run()` that runs
after the Check-mode early return and before stealth setup:
it calls `ensureGitignore` and `ensureOverrideStubs`, prints
the summary, and returns -- the asset walk, plugin
activation, sub-tools, AGENTS.md sections, and migrations
never run. Conflict validation lives in the CLI layer
(`cmd/unbound-force/main.go`, `runInit`), failing before any
filesystem write.

The issue names three conflicts (`--divisor`, `--force`,
`--check`); this design adds `--stealth` as a fourth. Rationale:
stub seeding (D4) and the gitignore block are both skipped
under stealth, so `--overrides --stealth` would be a
guaranteed silent no-op -- an incoherent mode. Failing closed
with an error naming both flags is strictly better than a
silent no-op and does not contradict the issue's acceptance
criteria. Help text (Long description and the flag table in
`docs/cli-reference.md`) documents the mode and all four
exclusions.

### D6: gitignoreBlock sync with semantic grouping

The six patterns join the block in semantically grouped
sections rather than one lump: `.uf/sibling-clones/`,
`.uf/sibling-cache/`, `.uf/feedback/`, and `.uf/progress/`
form a workspace-state group beside the runtime-data group;
`.uf/config.yaml` and `.uf/dewey/config.yaml` join the
per-developer group, whose comment is pluralized to cover
configuration alongside the two override files. The block's
first line (marker) and final line (`.unbound-force/`, the
legacy-dirs terminator) are unchanged -- the terminator's
stability across old and new block versions is what makes
D7's in-place upgrade possible. `.gaze/` stays out (separate
hero).

### D7: Managed-block replacement via shared extraction

A single helper, `extractManagedRegion(content, marker,
terminator)`, returns the byte range of a managed block: the
first line equal to the marker through the first subsequent
line equal to the terminator. `ensureGitignore` uses it with
(`gitignoreMarker`, `.unbound-force/`); the SC-FR-013 drift
test uses the same helper against the repo's own
`.gitignore`, so the verification cannot drift from the
behavior it verifies (Constitution VI).

Disposition logic, in order: no marker -> append the block
(`created`); marker and region byte-identical to the canonical
block -> leave untouched (`unchanged`, today's "already
configured"); marker present, region differs -> rewrite the
region in place (`upgraded`), preserving all content before
and after, never duplicating the marker; marker present, no
terminator after it -> leave untouched (`skipped` with detail
naming the malformed region -- never guess a region boundary).
Rewrites use the injected `Options.WriteFile`, keeping the
function testable in isolation (Constitution IV).

### D8: Stealth exclude block gains a closing sentinel

The stealth exclude block currently has a start marker but no
terminator, so its region is not deterministically delimited
(extra untracked-path lines are appended directly after the
canonical paths). New blocks are written as start marker +
escaped canonical paths + a new closing sentinel constant, and
`writeGitInfoExclude` applies the D7 disposition logic to its
own region. `appendNewUntrackedToExclude` keeps appending at
end of file -- after the sentinel, outside the managed region
-- so extra lines trivially survive an in-place upgrade;
`stealthExcludedPaths` parses all non-comment lines regardless
of region, so extras stay counted; rollback restores the
pre-init exclude bytes wholesale, so it is unaffected. Legacy
blocks (marker, no sentinel) take the malformed-skip path:
safe (never corrupts), and acceptable because stealth is
local-only and the canonical path set is unchanged today, so
legacy blocks are content-current; deleting the block and
re-running recreates it with the sentinel.

### D9: `.opencode/.gitignore` from Go code, merge semantics

Go's `//go:embed` excludes dotfiles from directory matches, so
the assets tree cannot carry the file; it is written from a Go
string constant by an ensure-style writer. Canonical content:
`node_modules`, `bun.lock`, `package-lock.json`, `.vitest/`,
`coverage/` (the vitest v8 provider writes coverage to
`.opencode/coverage/` by default; `.vitest/` is the cache;
`package-lock.json` joined the canonical set through the
stop-shipping-plugin-lock reconciliation). The writer is a
sub-tool independent of the plugin-activation result (the
install may fail or be skipped while `node_modules/` can
still appear), runs in full and `--divisor` mode (both run plugin
activation inside `.opencode/`), and is skipped under
`--stealth` (the wholesale `.opencode/` exclusion already
keeps the tree clean) and in `--overrides` mode. Merge
semantics: file absent -> write the canonical lines; file
present -> append only canonical lines missing from it,
retaining every existing user line in place. The repo's
canonical `.opencode/.gitignore` is updated to the five
canonical lines and pinned by the
`TestPluginPackageInputs_GitignoreTracksCanonicalInputs`
want-string (scaffold_test.go:1616); the now-redundant root
`.opencode/.vitest/` and `coverage/` lines are dropped from
the repo `.gitignore` (the scoped file covers them).

### Amendment of prior decisions

- **SC-FR-010** (model-matrix-scaffold-deployment, issue
  #889): the clause "*`*-custom.md` packs and the gitignored
  `.uf/model-matrix.override.yaml` MUST remain preserved and
  never scaffolded beyond the gitignore entry*" is modified by
  this change's delta to split the clause -- custom packs stay
  preserved user-owned files, while the override file is
  seeded once as a create-only-if-absent comment-only stub and
  is never modified afterwards, not even by `--force`.
- **D10** (model-matrix-delegation, issue #887): the ownership
  sentence "*the `.uf/model-matrix.override.yaml` file is
  user-owned -- gitignored, never scaffolded beyond the
  gitignore entry, exempt from drift checks exactly like
  `*-custom.md` packs*" is amended to read: user-owned,
  gitignored, seeded once as a create-only-if-absent
  comment-only stub, never modified afterwards (not even by
  `--force`), and exempt from drift checks (there is no
  canonical tracked source to compare against).
- **Gitignore capability idempotence clause** (archived
  change 2026-04-28-gitignore-init, `specs/gitignore.md`
  lines 15-17 and scenario lines 38-43): the clause "*if
  the marker comment `# Unbound Force — managed by uf init`
  is already present, the function MUST skip the append*"
  (and its ".gitignore is not modified (idempotent)"
  scenario) is superseded by this change's SC-FR-014: a
  present marker with a differing managed block now
  upgrades the region in place, deterministically bounded
  by the terminator; a malformed region (marker without
  terminator) still skips. The superseded record was never
  synced into the openspec/specs/ live tree and already
  lists removed patterns (`.uf/workflows/`,
  `.uf/muti-mind/artifacts/`, `.uf/mx-f/data/`,
  `.muti-mind/`, `.mx-f/`); the reconciliation lives here
  rather than a MODIFIED delta because the prior contract
  is a root Speckit spec outside OpenSpec's delta
  machinery.

Rationale for the reversal: the never-scaffolded stance
protected the per-developer file from team-init churn.
Create-only-if-absent preserves that protection entirely --
the file is only ever written when it does not exist -- while
removing the onboarding friction of hand-creating the
override from the docs. The `--force` carve-out follows from
the same principle: `--force` is the destructive reset for
tracked canonical files, and a per-developer file is never
canonical. Every run reports what it did to each stub
(Constitution III), so the seeding is observable, never
silent.

## Risks / Trade-offs

- **Managed-region rewrites clobber user edits inside the
  region** -- by design: the region between the marker and the
  terminator is tool-managed (the marker says so), and an
  upgrade restores it to canonical content. Users who need
  their own patterns must place them outside the region. The
  malformed-skip path refuses to guess when the terminator is
  ambiguous, so the rewrite never mangles a file it does not
  fully understand.
- **Deleting an override and re-running init re-seeds it** --
  a developer who deliberately deleted the override to force
  host-model defaults gets the comment-only stub back on the
  next init. Harmless (comment-only is a no-op) and
  consistent with create-only-if-absent.
- **Legacy stealth blocks never upgrade** (D8) -- accepted:
  safe skip, local-only surface, re-creatable.
- **`--stealth` added to the conflict set beyond the issue's
  three** (D5) -- a strengthening, not a contradiction; the
  alternative is a silent no-op mode.
- **The upgrade path rewrites a tracked file (`.gitignore`)**:
  on init re-runs in existing repos, that is the healing
  this change exists to deliver (issue #902), it is reported
  per run, and the drift test keeps the canonical block and
  the repo's own block identical going forward so the rewrite
  converges to a no-op.
- **No loader involvement** -- the stub content is pinned by
  exact-content tests in the scaffold suite, and the loaders'
  comment-only no-op behavior is already pinned by their own
  contracts; a change to either surface fails independently.
