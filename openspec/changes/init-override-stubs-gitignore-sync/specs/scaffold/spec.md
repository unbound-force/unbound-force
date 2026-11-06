<!-- capability: scaffold
     change: init-override-stubs-gitignore-sync
     originating_issue: 903 -->

Requirement numbering: SC-FR-001 through SC-FR-006 are allocated
by prior changes (review-council-multi-model-fanout and
successors); SC-FR-007 by model-matrix-delegation (slice 1,
issue #887); SC-FR-008 through SC-FR-010 by
model-matrix-scaffold-deployment (issue #889). SC-FR-011
through SC-FR-015 are allocated here; SC-FR-016 through
SC-FR-018 remain reserved for future scaffold changes. Sibling
capabilities: matrix loading and selection semantics live
in the unarchived model-matrix-delegation change delta,
openspec/changes/model-matrix-delegation/specs/
model-matrix/spec.md (MM-FR-001 through MM-FR-008,
unchanged); review-dispatch override semantics live in the
review-matrix capability (RD-FR-020, unchanged -- the
comment-only no-op behavior SC-FR-011 relies on is already
specified there).

## ADDED Requirements

### Requirement: Override Stub Seeding

**ID: SC-FR-011** This requirement MUST be satisfied.

A full-mode `uf init` run MUST seed both per-developer override
files -- `.uf/model-matrix.override.yaml` and
`.uf/review-matrix.override.yaml` -- as comment-only stub
templates, create-only-if-absent:

- each stub MUST be defined as a Go string constant in
  `internal/scaffold/scaffold.go` (the `gitignoreBlock`
  precedent), not an embedded asset: the files are gitignored
  per-developer data with no canonical tracked source for
  drift comparison, and Go constants avoid the embedded-asset
  registration checklist;
- each stub's content MUST consist solely of YAML comments
  documenting the file's whitelisted override paths with
  commented-out examples (model-matrix: `profiles.*.model`,
  `profiles.*.variant`, `classifier.llm.*`; review-matrix:
  `profiles.<tier>.model` and `profiles.<tier>.variant`), so
  the seeded file parses to YAML null and loads as a valid
  no-op in both existing loaders without any loader change;
- seeding MUST be create-only-if-absent: an existing override
  file, whatever its content, MUST be preserved byte-for-byte
  on every subsequent `uf init` run, including runs with
  `--force` -- per-developer data is never reset by the
  scaffold, which resets tracked canonical files only;
- seeding MUST be skipped under `--stealth` (stealth is a
  local-only trial whose rollback sweeps `.uf/`;
  per-developer state is not seeded into it) and in
  `--divisor` mode (a minimal review-council subset
  deployment into possibly-foreign repositories does not
  write per-developer files);
- each stub MUST report a per-file disposition (created or
  skipped) through the existing sub-tool result summary.

#### Scenario: Fresh full scaffold creates both stubs

- **GIVEN** an empty target directory
- **WHEN** `uf init` runs in full (non-DivisorOnly,
  non-stealth, non-OverridesOnly) mode
- **THEN** both `.uf/model-matrix.override.yaml` and
  `.uf/review-matrix.override.yaml` are created with exactly
  their stub-constant content and appear in the run's created
  results

#### Scenario: Existing override survives a plain re-run

- **GIVEN** a scaffolded repository whose
  `.uf/model-matrix.override.yaml` was hand-edited after
  deployment
- **WHEN** `uf init` re-runs without `--force`
- **THEN** the override file is byte-for-byte unchanged and
  reported as skipped

#### Scenario: Existing override survives --force

- **GIVEN** a scaffolded repository with a hand-edited
  `.uf/review-matrix.override.yaml`
- **WHEN** `uf init` re-runs with `--force`
- **THEN** tracked canonical files are replaced per the
  existing `--force` contract while the override file is
  byte-for-byte unchanged

#### Scenario: Seeded stub loads as a no-op

- **GIVEN** a freshly seeded, comment-only
  `.uf/model-matrix.override.yaml` next to a valid tracked
  matrix
- **WHEN** the matrix pair is loaded
- **THEN** the base matrix is used without modification (the
  comment-only file parses to YAML null)

#### Scenario: Stealth mode writes no override files

- **GIVEN** an empty target directory inside a git repository
- **WHEN** `uf init --stealth` runs
- **THEN** neither override file exists afterward

#### Scenario: Divisor mode writes no override files

- **GIVEN** an empty target directory
- **WHEN** `uf init --divisor` runs
- **THEN** neither override file exists afterward

### Requirement: Overrides-Only Init Mode

**ID: SC-FR-012** This requirement MUST be satisfied.

`uf init --overrides` MUST create only the two override stubs
(SC-FR-011) plus the managed `.gitignore` block
(`ensureGitignore`):

- the mode MUST skip the embedded-asset walk, sub-tool
  initialization, AGENTS.md sections, and directory migrations
  entirely -- no other file or directory is created or
  modified;
- the mode MUST be idempotent: re-running it with both stubs
  present and an intact managed block modifies nothing;
- the flag MUST be mutually exclusive with `--divisor`,
  `--force`, `--check`, and `--stealth`; a run combining
  `--overrides` with any of them MUST fail flag validation in
  `cmd/unbound-force/main.go` with an error naming the
  conflicting flags, before any filesystem write;
- the command help text MUST document the mode and its
  exclusions.

#### Scenario: Overrides-only creates exactly stubs and block

- **GIVEN** an empty target directory
- **WHEN** `uf init --overrides` runs
- **THEN** the run creates exactly
  `.uf/model-matrix.override.yaml`,
  `.uf/review-matrix.override.yaml`, and a `.gitignore`
  containing the managed block, and creates no other file or
  directory (no `.opencode/`, `.specify/`, `openspec/`, or
  `.uf/` content beyond the two stubs)

#### Scenario: Overrides-only is idempotent

- **GIVEN** a directory where a prior `uf init --overrides`
  created both stubs and the managed block
- **WHEN** `uf init --overrides` runs again
- **THEN** no file is modified and every disposition reports
  skipped or unchanged

#### Scenario: Conflicting flags fail before any write

- **GIVEN** `uf init --overrides` combined with `--divisor`,
  `--force`, `--check`, or `--stealth`
- **WHEN** the command runs
- **THEN** it exits non-zero with an error naming
  `--overrides` and the conflicting flag, and the target
  directory is untouched

#### Scenario: Existing override preserved in overrides-only

- **GIVEN** a directory where `.uf/model-matrix.override.yaml`
  already exists with hand-edited content
- **WHEN** `uf init --overrides` runs
- **THEN** only `.uf/review-matrix.override.yaml` and, if
  absent, the managed `.gitignore` block are created, and the
  existing override file is byte-for-byte unchanged

### Requirement: Gitignore Block Pattern Currency

**ID: SC-FR-013** This requirement MUST be satisfied.

The scaffolded managed gitignore block MUST stay current with
the runtime and per-developer paths Unbound Force tools create
in target repositories:

- `gitignoreBlock` MUST include `.uf/config.yaml`,
  `.uf/dewey/config.yaml`, `.uf/sibling-clones/`,
  `.uf/sibling-cache/`, `.uf/feedback/`, and `.uf/progress/`,
  grouped semantically with the existing `.uf/` sections;
- `gitignoreBlock` MUST NOT include `.gaze/` (gaze is a
  separate hero not deployed by `uf init`);
- a scaffold drift test MUST assert that this repository's own
  managed block -- the marker line through the
  `.unbound-force/` terminator line in the repo root
  `.gitignore` -- equals `gitignoreBlock` byte-for-byte, using
  the same managed-region extraction helper the replacement
  logic (SC-FR-014) uses, so scaffold-to-repo drift fails the
  suite naming the file;
- the repo root `.gitignore` managed region MUST contain no
  duplicated pattern (the `.uf/dewey/config.yaml` duplicate is
  removed) and its per-developer override comment MUST name
  both override files.

#### Scenario: Fresh target ignores all runtime paths

- **GIVEN** an empty target directory
- **WHEN** `uf init` runs and appends the managed block
- **THEN** the block contains `.uf/config.yaml`,
  `.uf/dewey/config.yaml`, `.uf/sibling-clones/`,
  `.uf/sibling-cache/`, `.uf/feedback/`, and `.uf/progress/`

#### Scenario: Gaze directory stays out of the block

- **GIVEN** the canonical `gitignoreBlock`
- **WHEN** its patterns are enumerated
- **THEN** `.gaze/` is not among them

#### Scenario: Sibling section no longer contradicts the block

- **GIVEN** a repository scaffolded by `uf init`
- **WHEN** the scaffolded AGENTS.md sibling-repositories
  section states that clones and caches live under
  `.uf/sibling-clones/` and `.uf/sibling-cache/` as gitignored
  paths
- **THEN** the deployed managed block ignores both paths

#### Scenario: Drift between repo and scaffold fails the suite

- **GIVEN** the repo root `.gitignore` managed region differs
  from `gitignoreBlock` by one pattern
- **WHEN** the drift test runs
- **THEN** it fails naming `.gitignore` as the drifted file

### Requirement: Managed Block Replacement Upgrade

**ID: SC-FR-014** This requirement MUST be satisfied.

`ensureGitignore()` and the stealth-mode `.git/info/exclude`
writer MUST upgrade stale managed blocks in place instead of
leaving them behind:

- when the managed marker is present and the extracted
  managed region is byte-identical to the canonical block, the
  file MUST be left unchanged and the run MUST report the
  unchanged disposition;
- when the managed marker is present but the extracted region
  differs from the canonical block, the region MUST be
  rewritten in place with the canonical block --
  deterministically, without creating duplicate blocks, and
  preserving all content before and after the region -- and
  the run MUST report the upgraded disposition;
- when the marker is present but the region is malformed (no
  terminator line after the marker), the file MUST be left
  unchanged and the run MUST report a skipped disposition with
  detail naming the malformed region;
- when no marker is present, the block MUST be appended (the
  existing creation path) and the run MUST report the created
  disposition;
- the `.gitignore` managed region MUST be delimited by the
  marker line and the `.unbound-force/` terminator line (the
  final line of `gitignoreBlock`); region extraction MUST be
  performed by a single helper shared with the SC-FR-013 drift
  test;
- the stealth exclude block MUST gain a closing sentinel line
  so its region is deterministically delimited the same way,
  and its writer MUST apply the identical created / upgraded /
  unchanged / malformed-skip semantics to its own
  marker-delimited region; extra untracked-path lines appended
  after the region MUST survive an in-place upgrade.

#### Scenario: Stale gitignore block upgrades in place

- **GIVEN** a `.gitignore` containing the managed marker, an
  out-of-date block body, and trailing user content after the
  terminator line
- **WHEN** `uf init` runs
- **THEN** the marker-to-terminator region is replaced with
  the canonical `gitignoreBlock`, the trailing user content is
  preserved, the file contains exactly one managed marker, and
  the run reports the `.gitignore` disposition as upgraded

#### Scenario: Current block is left untouched

- **GIVEN** a `.gitignore` whose managed region is
  byte-identical to `gitignoreBlock`
- **WHEN** `uf init` runs
- **THEN** the file is byte-for-byte unchanged and the
  disposition reports unchanged

#### Scenario: Malformed block is skipped with detail

- **GIVEN** a `.gitignore` containing the managed marker with
  no `.unbound-force/` terminator line after it
- **WHEN** `uf init` runs
- **THEN** the file is unchanged and the disposition reports
  skipped with a detail naming the malformed managed region

#### Scenario: Stealth exclude block upgrades the same way

- **GIVEN** a `.git/info/exclude` whose stealth marker-delimited
  region is stale and whose region is followed by
  extra untracked-path lines
- **WHEN** `uf init --stealth` runs
- **THEN** the region is rewritten with the canonical paths
  and closing sentinel, the extra lines after the region are
  preserved, and the run reports the exclude disposition as
  upgraded

#### Scenario: Legacy stealth block without sentinel is skipped

- **GIVEN** a `.git/info/exclude` written by an older `uf init`
  that has the stealth marker but no closing sentinel line
- **WHEN** `uf init --stealth` runs
- **THEN** the file is unchanged and the disposition reports
  skipped with detail (the malformed-region path)

### Requirement: OpenCode Directory Ignore Sub-Tool

**ID: SC-FR-015** This requirement MUST be satisfied.

The scaffold MUST deploy `.opencode/.gitignore` so that
plugin-dependency and test-output noise inside `.opencode/` is
ignored in fresh targets:

- the file MUST be written from Go code (a Go string constant
  in the scaffold engine) because Go's `//go:embed` excludes
  dotfiles from directory matches and the embedded assets tree
  cannot carry it;
- the canonical `.opencode/.gitignore` in this repository MUST
  contain exactly `node_modules`, `bun.lock`,
  `package-lock.json`, `.vitest/`, and `coverage/` (one per
  line; `package-lock.json` joined the set through the
  stop-shipping-plugin-lock reconciliation), and the deployed
  content MUST equal the canonical content, pinned by the
  existing canonical-inputs test's want-string;
- the sub-tool MUST run independently of the review-plugin
  activation result (the install may fail or be skipped while
  `node_modules/` can still appear);
- the writer MUST be idempotent and MUST preserve unknown user
  lines: when the file exists, only canonical lines missing
  from it are appended, and existing user lines are retained
  in place;
- deployment MUST run in full mode and `--divisor` mode (both
  run plugin activation inside `.opencode/`) and MUST be
  skipped under `--stealth` (the wholesale `.opencode/`
  exclusion already keeps the tree clean there) and in
  `--overrides` mode (SC-FR-012 restricts that mode to the
  stubs and the managed block).

#### Scenario: Fresh target gets the opencode ignore file

- **GIVEN** an empty target directory
- **WHEN** `uf init` runs in full mode
- **THEN** `.opencode/.gitignore` is created with exactly the
  canonical five lines

#### Scenario: User lines preserved, canonical lines added

- **GIVEN** an existing `.opencode/.gitignore` containing a
  user line `foo.log` and the canonical line `node_modules`
- **WHEN** `uf init` runs
- **THEN** `foo.log` and `node_modules` are retained, the
  other canonical lines are appended, and no user line is
  removed

#### Scenario: Deployment survives plugin-activation failure

- **GIVEN** a target where review-plugin activation fails
  (for example, install errors)
- **WHEN** `uf init` runs
- **THEN** `.opencode/.gitignore` is still written and its
  disposition is reported

## MODIFIED Requirements

### Requirement: Deployed Matrix Re-Init Ownership

**ID: SC-FR-010** This requirement MUST be satisfied.

The deployed `.uf/model-matrix.yaml` MUST be a project-owned
extension point, mirroring the `uf/review-matrix.yaml` and
`uf/reviewer-capabilities.yaml` init-survival precedent
(issues #884/#886):

- a fresh `uf init` run MUST seed `.uf/model-matrix.yaml` from
  the embedded template (the new-file creation path);
- a plain `uf init` re-run MUST NOT overwrite a deployed
  `.uf/model-matrix.yaml` whose content differs from the
  template -- `isToolOwned("uf/model-matrix.yaml")` MUST return
  false, so team-edited tracked matrices survive re-init, and
  `--force` remains the explicit path that replaces them;
- `.uf/schemas/model-matrix/` files and canonical convention
  packs (including `meta-prompt.md`) MUST remain tool-owned
  and refresh on re-run when content differs;
- `*-custom.md` packs MUST remain preserved user-owned files;
- the gitignored `.uf/model-matrix.override.yaml` MUST be
  seeded once as a create-only-if-absent comment-only stub
  (SC-FR-011) and MUST NOT be modified by any subsequent
  `uf init` run -- not even by `--force`, which resets tracked
  canonical files only.

Previously: `*-custom.md` packs and the gitignored
`.uf/model-matrix.override.yaml` MUST remain preserved and
never scaffolded beyond the gitignore entry. The reversal
rationale: the never-scaffolded stance protected the
per-developer file from team-init churn; create-only-if-absent
seeding preserves that protection entirely -- the file is only
ever written when it does not exist -- while removing the
onboarding friction of hand-creating the override from the
docs.

#### Scenario: Team-edited matrix survives re-init

- **GIVEN** a repo scaffolded by `uf init` whose tracked
  `.uf/model-matrix.yaml` was edited by the team after
  deployment
- **WHEN** `uf init` re-runs without `--force`
- **THEN** the tracked matrix appears in the run's skipped
  list with its edited content unchanged

#### Scenario: Tool-owned matrix files refresh on diff

- **GIVEN** a scaffolded repo whose deployed
  `.uf/schemas/model-matrix/v1.schema.json` (or canonical
  `meta-prompt.md`) differs from the embedded copy
- **WHEN** `uf init` re-runs without `--force`
- **THEN** the differing file appears in the run's updated
  list and matches the embedded content afterward

#### Scenario: Identical content produces no spurious updates

- **GIVEN** a scaffolded repo whose tool-owned matrix files
  are byte-identical to the embedded copies
- **WHEN** `uf init` re-runs without `--force`
- **THEN** those files appear in the skipped list, not the
  updated list

#### Scenario: Force replaces the team-edited matrix

- **GIVEN** a scaffolded repo with a team-edited tracked
  `.uf/model-matrix.yaml`
- **WHEN** `uf init` re-runs with `--force`
- **THEN** the tracked matrix is overwritten with the embedded
  template content

#### Scenario: Force preserves the override file

- **GIVEN** a scaffolded repo whose
  `.uf/model-matrix.override.yaml` was hand-edited after
  deployment
- **WHEN** `uf init` re-runs with `--force`
- **THEN** the tracked matrix is replaced per the scenario
  above while the override file is byte-for-byte unchanged

## REMOVED Requirements

None. No scaffold requirement is removed; SC-FR-010's override
clause is modified in place, and every prior scenario of the
modified requirement remains satisfied.
