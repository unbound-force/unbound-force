## Why

Speckit branches currently use a flat `NNN-<name>` naming
convention (e.g., `001-org-constitution`, `031-unleash-openspec`),
while OpenSpec branches are organized under the `opsx/` folder
prefix (e.g., `opsx/sandbox-uid-mapping`). This inconsistency
means Speckit branches pollute the top-level branch namespace,
making `git branch` output harder to scan, and the two workflow
types are not visually distinguishable at a glance.

Upstream Speckit (github/spec-kit) has open issues requesting
custom branch namespacing (#1382, #3081) but no resolution yet.

Fixes: https://github.com/unbound-force/unbound-force/issues/149

## What Changes

Add a `speckit/` folder prefix to all newly created Speckit
feature branches. The branch naming convention changes from
`NNN-<name>` to `speckit/NNN-<name>`.

All branch-detection logic becomes backward-compatible,
accepting both the legacy `NNN-<name>` pattern and the new
`speckit/NNN-<name>` pattern. Existing branches are not
renamed.

The `specs/` directory structure on disk is unchanged --
spec artifacts continue to live at `specs/NNN-<name>/`.

## Capabilities

### New Capabilities
- `speckit/ branch prefix`: New Speckit branches are created
  as `speckit/NNN-<name>`, organizing them under a folder in
  the git branch namespace

### Modified Capabilities
- `Branch detection`: All commands and skills that detect
  Speckit branches (unleash, review-council, finale,
  address-feedback, agent-brief, cobalt-crush, review-context,
  speckit-workflow) accept both `NNN-<name>` and
  `speckit/NNN-<name>` patterns
- `Prefix stripping`: Branch-to-directory mapping strips the
  `speckit/` prefix before constructing filesystem paths to
  `specs/NNN-<name>/`

### Removed Capabilities
- None

## Impact

- **Shell scripts** (2 files): `create-new-feature.sh` branch
  name construction, `common.sh` validation regexes and
  branch-to-directory mapping
- **Commands** (8 files): Branch detection patterns and error
  messages in unleash, review-council, finale,
  address-feedback, agent-brief, cobalt-crush, and workflow
  status/advance commands
- **Skills** (1 file): review-context branch-to-spec mapping
  table (speckit-workflow only has filesystem paths, unchanged)
- **Scaffold assets** (7 files): Mirrors of updated commands
  and skills under `internal/scaffold/assets/`
- **Documentation** (3 files): AGENTS.md, docs/usage.md,
  docs/architecture.md
- **Total**: 21 modified files + 4 new OpenSpec artifacts

Note: Initial estimate was ~56 files / ~100 edits. Actual
scope is smaller because Speckit command guardrails (11
files), agent file examples (7 files), constitution, and
doctor_test.go only reference `specs/NNN-*/` filesystem
paths, not branch names, so they required no changes.

Historical spec artifacts and archived OpenSpec changes are
NOT modified.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

This change affects branch naming conventions only. All
artifact-based communication (spec files, reports, schemas)
remains unchanged. Spec artifacts continue to be written to
`specs/NNN-<name>/` on disk. No inter-hero communication
patterns are affected.

### II. Composability First

**Assessment**: PASS

The change is internal to the Unbound Force meta-repository.
No hero's standalone functionality is affected. The backward-
compatible detection logic ensures that existing branches
created by any hero continue to work without modification.

### III. Observable Quality

**Assessment**: N/A

This change does not affect machine-parseable output, artifact
formats, or provenance metadata. It is a developer-workflow
convention change.

### IV. Testability

**Assessment**: PASS

The backward-compatible regex patterns are testable in
isolation. The `internal/doctor/checks.go` directory-name
regex (`^\d{3}-`) scans filesystem entries, not branch names,
and requires no change. Test fixtures in `doctor_test.go` that
reference branch name patterns will be updated.

### V. Security by Default

**Assessment**: N/A

No new dependencies, inputs, or privilege changes. Branch
naming is a developer convention with no security surface.
