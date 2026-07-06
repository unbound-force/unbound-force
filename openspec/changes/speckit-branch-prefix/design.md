## Context

Speckit branches use flat `NNN-<name>` names while OpenSpec
branches use the `opsx/` folder prefix. Issue #149 requests
parity. The proposal selects `speckit/` as the prefix after
evaluating `spec/` (confusable with `specs/` directory via
singular/plural), `specs/` (double-path risk in shell scripts
that do `$SPECS_DIR/$BRANCH_NAME`), and `speckit/` (zero
ambiguity, fails fast if prefix strip is missed).

The shell scripts in `.specify/scripts/bash/common.sh`
interpolate branch names directly into filesystem paths via
`get_feature_dir()` and `find_feature_dir_by_prefix()`. Any
prefix that collides with or is similar to the `specs/`
directory creates a subtle double-path bug (`specs/specs/...`).
`speckit/` avoids this entirely -- a missed strip produces
`specs/speckit/001-feature/`, which fails fast and obviously.

## Goals / Non-Goals

### Goals
- New Speckit branches are created as `speckit/NNN-<name>`
- All branch-detection logic accepts both `NNN-<name>`
  (legacy) and `speckit/NNN-<name>` (new)
- Branch-to-directory mapping correctly strips the `speckit/`
  prefix before constructing `specs/NNN-<name>/` paths
- Documentation, commands, skills, agents, and scaffold assets
  reflect the new convention
- Upstream Speckit issue #1382 receives a comment proposing
  `branch_prefix` as a config key in the git extension

### Non-Goals
- Renaming existing remote branches (they remain as-is)
- Modifying historical spec artifacts or archived OpenSpec
  changes
- Changing the `specs/` directory structure on disk
- Modifying the OpenSpec `opsx/` convention
- Submitting an upstream PR (comment/proposal only)
- Adding a configurable prefix mechanism locally (hardcode
  `speckit/` for now; migrate to config-driven if upstream
  adds `branch_prefix` support)

## Decisions

### D1: Prefix is `speckit/` (not `spec/` or `specs/`)

**Rationale**: `spec/` is confusable with the `specs/`
directory (singular vs plural). `specs/` causes a double-path
bug in `common.sh` where `get_feature_dir()` does
`$repo_root/specs/$branch_name` -- if the branch is
`specs/001-feature`, the result is
`$repo_root/specs/specs/001-feature`. `speckit/` has zero
collision risk: a missed prefix strip produces
`specs/speckit/001-feature/` which fails fast and obviously.

### D2: Backward-compatible detection (accept both patterns)

**Rationale**: Existing branches should continue to work.
Detection regexes change from `^[0-9]{3}-` to
`^(speckit/)?[0-9]{3}-` in shell scripts, and from `NNN-*`
to `speckit/NNN-*` (with legacy fallback) in command/skill
prose. This is a non-breaking change.

### D3: Prefix stripping in common.sh

The `find_feature_dir_by_prefix()` function and
`get_current_branch()` must strip the `speckit/` prefix
before path construction. The strip happens early -- at the
point where the branch name enters the path-construction
pipeline -- rather than at each consumer site.

Implementation:
```bash
# In get_current_branch() or find_feature_dir_by_prefix():
branch_name="${branch_name#speckit/}"
```

This uses bash parameter expansion to strip the prefix only
if present, making it safe for both old and new branch names.

### D4: Spec directory names on disk are unchanged

Branch `speckit/001-feature` maps to directory
`specs/001-feature/`. The `create-new-feature.sh` script
sets `BRANCH_NAME="speckit/${FEATURE_NUM}-${BRANCH_SUFFIX}"`
for git but uses `FEATURE_DIR="$SPECS_DIR/${FEATURE_NUM}-${BRANCH_SUFFIX}"`
(without the prefix) for the filesystem directory.

### D5: Historical artifacts are not modified

Completed specs in `specs/016-*`, `specs/018-*`,
`specs/030-*`, `specs/031-*` and archived OpenSpec changes
document conventions as they were at time of writing.
Modifying them would be revisionist and create unnecessary
churn.

### D6: `internal/doctor/checks.go` is unchanged

The `hasSpecNumberedDirs()` function's regex `^\d{3}-` scans
immediate children of the `specs/` directory, not branch
names. Spec directories continue to be named `NNN-<name>/`
on disk, so no change is needed.

## Risks / Trade-offs

### R1: Upstream divergence (LOW)

The change modifies `.specify/scripts/bash/create-new-feature.sh`
and `common.sh`, which are scaffolded from upstream Speckit.
This increases divergence. Mitigated by:
- The change is isolated to 2 local scripts
- If upstream adds `branch_prefix` config support, we migrate
  to the config-driven approach and drop local overrides
- A comment on upstream #1382 proposes this feature

### R2: Missed pattern in a command file (LOW)

With ~100 edits across ~56 files, it is possible to miss a
pattern. Mitigated by:
- Backward-compatible regex means old patterns still work
- A grep sweep in the verification phase catches stragglers
- Scaffold drift detection tests catch mismatches between
  live files and `internal/scaffold/assets/` copies

### R3: Branch name with `/` in shell scripts (LOW)

Git handles `/` in branch names natively (it creates
subdirectories under `.git/refs/heads/`). The `speckit/`
prefix is no different from how `opsx/` already works.
No special handling needed.
