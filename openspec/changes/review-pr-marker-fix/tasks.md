<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Sub-Agent Prompt Exception

- [x] 1.1 Read `.opencode/commands/uf.review-pr.md` in full. Identify the exact location in the sub-agent prompt (Step F.3 Constitution Compliance, or a new pre-filter clause immediately before it) where the marker exception will be added.
- [x] 1.2 Add a narrow exception clause to the sub-agent prompt that instructs: the sub-agent MUST NOT flag `<!-- code-review: passed -->` or `<!-- spec-review: passed -->` as gatekeeping violations when they appear in files matching `openspec/changes/*/tasks.md` or `specs/*/tasks.md` (relative to repo root). The exception MUST match both conditions (exact marker string AND exact path pattern) simultaneously. The sub-agent MUST continue to flag these markers in any other file path, and MUST continue to flag any other marker-like HTML comments or real gate weakening in task files.
- [x] 1.3 Verify the exception clause does not alter any other part of the sub-agent prompt. Verify the gatekeeping rules in AGENTS.md are unchanged.

## 2. Regression Test

- [x] 2.1 Create a regression test that reproduces the false-positive scenario from GitHub issue #539. The test MUST provide a synthetic diff containing a `tasks.md` file at a legitimate path (e.g., `openspec/changes/test-feature/tasks.md`) with the marker `<!-- code-review: passed -->` and assert that no gatekeeping violation finding is produced.
- [x] 2.2 The regression test MUST include a negative control: assert that a non-standard marker (e.g., `<!-- coverage-threshold: lowered -->`) in the same file path IS still flagged as a gatekeeping violation.
- [x] 2.3 The regression test MUST include a path-negative control: assert that `<!-- code-review: passed -->` in a non-task file (e.g., `AGENTS.md`) IS still flagged as a gatekeeping violation.
- [x] 2.4 Run the full test suite (`make test` or `go test -race -count=1 ./...`) and verify all tests pass, including the new regression test.

## 3. Verification

- [x] 3.1 Run `make check` (lint + test + build) and verify all checks pass.
- [x] 3.2 Verify constitution alignment: confirm this change aligns with all five org constitution principles (I. Autonomous Collaboration, II. Composability First, III. Observable Quality, IV. Testability, V. Security by Default) as assessed in the proposal.
- [x] 3.3 Verify the change does NOT modify any gatekeeping gate values (coverage thresholds, CI flags, severity definitions, convention rules, constitution gates, pinned dependency versions) per AGENTS.md Behavioral Rules.
- [x] 3.4 Verify the change does NOT expand into acceptance-criteria-review scope (#563).

<!-- spec-review: passed -->

<!-- code-review: passed -->

<!-- scaffolded by uf vdev -->
