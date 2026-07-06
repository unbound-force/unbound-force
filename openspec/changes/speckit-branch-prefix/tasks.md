<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file --
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Core shell script logic (FR-001, FR-002, FR-003, FR-004)

- [x] 1.1 [P] Update `.specify/scripts/bash/create-new-feature.sh`:
  change branch name construction at line 251 from
  `BRANCH_NAME="${FEATURE_NUM}-${BRANCH_SUFFIX}"` to
  `BRANCH_NAME="speckit/${FEATURE_NUM}-${BRANCH_SUFFIX}"`.
  Ensure `FEATURE_DIR` continues to use
  `$SPECS_DIR/${FEATURE_NUM}-${BRANCH_SUFFIX}` (no prefix
  on filesystem). Update any branch-name regex patterns
  (lines 116-117) to accept `speckit/` prefix.
- [x] 1.2 [P] Update `.specify/scripts/bash/common.sh`:
  (a) In `get_current_branch()`, strip `speckit/` prefix
  using `branch="${branch#speckit/}"` before returning.
  (b) Update `validate_branch()` regex from `^[0-9]{3}-`
  to `^(speckit/)?[0-9]{3}-` (line 75).
  (c) Update `find_feature_dir_by_prefix()` to strip
  `speckit/` from `$branch_name` before regex matching
  and path construction.
  (d) Update error message example (line 77) to show
  `speckit/001-feature-name`.

## 2. Command branch detection (FR-002)

- [x] 2.1 [P] Update `.opencode/commands/unleash.md`:
  change `NNN-*` pattern references to
  `speckit/NNN-*` with backward-compatible detection.
  Update error messages and examples (lines 57, 72, 91,
  94, 287, 560, 640).
- [x] 2.2 [P] Update `.opencode/commands/review-council.md`:
  change branch classification pattern from `NNN-*` to
  `speckit/NNN-*` (lines 79, 95, 667, 668).
- [x] 2.3 [P] Update `.opencode/commands/finale.md`:
  change branch detection and error messages (lines 24,
  45, 247).
- [x] 2.4 [P] Update `.opencode/commands/address-feedback.md`:
  change spec artifact discovery pattern (line 140).
- [x] 2.5 [P] Update `.opencode/commands/agent-brief.md`:
  change detection logic and governance templates
  (lines 82, 201, 213, 223).
- [x] 2.6 [P] Update `.opencode/commands/cobalt-crush.md`:
  change branch requirement message (line 93).
- [x] 2.7 [P] Update `.opencode/commands/workflow-status.md`:
  change hardcoded example (line 111).
- [x] 2.8 [P] Update `.opencode/commands/workflow-advance.md`:
  change hardcoded example (line 90).

## 3. Skill updates (FR-002, FR-003)

- [x] 3.1 [P] Update
  `.opencode/skills/review-context/SKILL.md`:
  change branch-to-spec mapping table to show
  `speckit/NNN-<name>` pattern with prefix stripping
  (lines 41, 49, 236).
- [x] 3.2 [P] Update
  `.opencode/skills/speckit-workflow/SKILL.md`:
  change task file discovery path (line 22).
  No change needed — only contains filesystem path
  `specs/NNN-*/tasks.md`, not branch names.

## 4. Speckit command guardrail blocks

- [x] 4.1-4.11: All guardrail blocks verified. They
  reference `specs/NNN-*/` filesystem paths, not branch
  names. No changes needed — filesystem directory naming
  is unchanged.

## 5. Agent file examples

- [x] 5.1-5.7: All agent files verified. They reference
  `specs/NNN-feature/artifact.md` filesystem paths in
  finding templates, not branch names. No changes
  needed — filesystem directory naming is unchanged.

## 6. Scaffold asset synchronization

- [x] 6.1 Sync all `internal/scaffold/assets/opencode/`
  copies with their live counterparts updated in tasks
  2.x, 3.x, 4.x, and 5.x. Files to sync:
  `commands/unleash.md`, `commands/review-council.md`,
  `commands/finale.md`, `commands/address-feedback.md`,
  `commands/agent-brief.md`, `commands/cobalt-crush.md`,
  `commands/uf-init.md`,
  `skills/review-context/SKILL.md`,
  `skills/speckit-workflow/SKILL.md`,
  `agents/divisor-guard.md`,
  `agents/divisor-architect.md`,
  `agents/divisor-adversary.md`,
  `agents/divisor-testing.md`,
  `agents/divisor-sre.md`,
  `agents/divisor-curator.md`.
  Note: no [P] marker because this task depends on
  prior tasks completing first.

## 7. Top-level documentation

- [x] 7.1 [P] Update `AGENTS.md`: change branch naming
  convention Branches line (line 239).
- [x] 7.2 [P] `.specify/memory/constitution.md` verified.
  References `specs/NNN-*/` filesystem path in Phase
  Discipline rule. No change needed.
- [x] 7.3 [P] Update `docs/usage.md`: change branch
  pattern reference (line 139).
- [x] 7.4 [P] Update `docs/architecture.md`: change
  branch pattern reference (line 447).

## 8. Verification

- [x] 8.1 Run `make check` (build, lint, tests). Confirm
  scaffold drift detection tests pass. PASSED: 0 lint
  issues, all 20 test packages pass, build succeeds.
- [x] 8.2 Grep sweep: searched for remaining `NNN-<name>`
  branch references. All remaining references are either:
  (a) filesystem paths (`specs/NNN-*/`), (b) already
  updated to show `speckit/NNN-*` with legacy fallback,
  or (c) in excluded historical files.
- [x] 8.3 Constitution alignment verified: I=PASS (no
  artifact format changes), II=PASS (no hero dependency
  changes), III=N/A (no output format changes), IV=PASS
  (all tests pass), V=N/A (no security surface changes).

<!-- spec-review: passed -->
