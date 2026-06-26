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

## 1. Rename Embedded Assets

Rename the 10 command files under
`internal/scaffold/assets/opencode/commands/`.
These are the canonical source files compiled into
the `uf` binary via `embed.FS`.

- [x] 1.1 [P] Rename `address-feedback.md` to
  `uf.address-feedback.md`
  Files: `internal/scaffold/assets/opencode/commands/address-feedback.md`

- [x] 1.2 [P] Rename `agent-brief.md` to
  `uf.agent-brief.md`
  Files: `internal/scaffold/assets/opencode/commands/agent-brief.md`

- [x] 1.3 [P] Rename `cobalt-crush.md` to
  `uf.cobalt-crush.md`
  Files: `internal/scaffold/assets/opencode/commands/cobalt-crush.md`

- [x] 1.4 [P] Rename `constitution-check.md` to
  `uf.constitution-check.md`
  Files: `internal/scaffold/assets/opencode/commands/constitution-check.md`

- [x] 1.5 [P] Rename `finale.md` to `uf.finale.md`
  Files: `internal/scaffold/assets/opencode/commands/finale.md`

- [x] 1.6 [P] Rename `review-council.md` to
  `uf.review-council.md`
  Files: `internal/scaffold/assets/opencode/commands/review-council.md`

- [x] 1.7 [P] Rename `review-pr.md` to
  `uf.review-pr.md`
  Files: `internal/scaffold/assets/opencode/commands/review-pr.md`

- [x] 1.8 [P] Rename `triage-issue.md` to
  `uf.triage-issue.md`
  Files: `internal/scaffold/assets/opencode/commands/triage-issue.md`

- [x] 1.9 [P] Rename `uf-init.md` to `uf.init.md`
  Files: `internal/scaffold/assets/opencode/commands/uf-init.md`

- [x] 1.10 [P] Rename `unleash.md` to
  `uf.unleash.md`
  Files: `internal/scaffold/assets/opencode/commands/unleash.md`

## 2. Rename Live Command Files

Rename the 10 command files under
`.opencode/commands/` (the deployed copies in this
repo).

- [x] 2.1 [P] Rename `address-feedback.md` to
  `uf.address-feedback.md`
  Files: `.opencode/commands/address-feedback.md`

- [x] 2.2 [P] Rename `agent-brief.md` to
  `uf.agent-brief.md`
  Files: `.opencode/commands/agent-brief.md`

- [x] 2.3 [P] Rename `cobalt-crush.md` to
  `uf.cobalt-crush.md`
  Files: `.opencode/commands/cobalt-crush.md`

- [x] 2.4 [P] Rename `constitution-check.md` to
  `uf.constitution-check.md`
  Files: `.opencode/commands/constitution-check.md`

- [x] 2.5 [P] Rename `finale.md` to `uf.finale.md`
  Files: `.opencode/commands/finale.md`

- [x] 2.6 [P] Rename `review-council.md` to
  `uf.review-council.md`
  Files: `.opencode/commands/review-council.md`

- [x] 2.7 [P] Rename `review-pr.md` to
  `uf.review-pr.md`
  Files: `.opencode/commands/review-pr.md`

- [x] 2.8 [P] Rename `triage-issue.md` to
  `uf.triage-issue.md`
  Files: `.opencode/commands/triage-issue.md`

- [x] 2.9 [P] Rename `uf-init.md` to `uf.init.md`
  Files: `.opencode/commands/uf-init.md`

- [x] 2.10 [P] Rename `unleash.md` to
  `uf.unleash.md`
  Files: `.opencode/commands/unleash.md`

## 3. Update Go Source — Scaffold Engine

Update hard references in `internal/scaffold/`.

- [x] 3.1 Add `renamedCommands` migration map to
  `scaffold.go` (near `isToolOwned`, ~line 302).
  Map old paths to new paths for all 10 commands.
  Files: `internal/scaffold/scaffold.go`

- [x] 3.2 Add orphan cleanup logic to `Run()` in
  `scaffold.go`. After the main scaffold walk,
  iterate `renamedCommands` and remove old-path
  files in the target directory. Report removed
  files in the summary.
  Files: `internal/scaffold/scaffold.go`

- [x] 3.3 Update `isDivisorAsset()` path check
  from `"opencode/commands/review-council.md"` to
  `"opencode/commands/uf.review-council.md"`
  (line 338).
  Files: `internal/scaffold/scaffold.go`

- [x] 3.4 Update `hintDivisor` constant from
  `"/review-council"` to `"/uf.review-council"`
  (line 1588).
  Files: `internal/scaffold/scaffold.go`

- [x] 3.5 Update warning message from `"/uf-init"`
  to `"/uf.init"` (line 920).
  Files: `internal/scaffold/scaffold.go`

## 4. Update Go Source — Doctor

Update hard references in `internal/doctor/`.

- [x] 4.1 Update all 7 `InstallHint` strings in
  `checks.go` from `"/agent-brief"` to
  `"/uf.agent-brief"`.
  Files: `internal/doctor/checks.go`

- [x] 4.2 [P] Update `InstallHint` assertion in
  `doctor_test.go` to expect `"/uf.agent-brief"`.
  Files: `internal/doctor/doctor_test.go`

## 5. Update Go Tests — Scaffold

Update test assertions in
`internal/scaffold/scaffold_test.go`.

- [x] 5.1 Update `expectedAssetPaths` list: replace
  all 10 old command paths with new `uf.*` paths.
  Files: `internal/scaffold/scaffold_test.go`

- [x] 5.2 Update `knownNonEmbeddedFiles` map:
  remove old-name entries, add new entries if
  needed (the renamed files are now embedded under
  new names; only non-embedded files belong here).
  Files: `internal/scaffold/scaffold_test.go`

- [x] 5.3 Update individual test assertions that
  reference old command filenames (20+ locations).
  Grep for old names and update systematically.
  Files: `internal/scaffold/scaffold_test.go`

- [x] 5.4 Add test for migration map orphan cleanup.
  Verify that when old-name command files exist in
  the target directory, `Run()` removes them after
  deploying new-name files. Cover the 3 spec
  scenarios: re-run migration, idempotent re-run,
  and fresh repository (no old files to remove).
  Files: `internal/scaffold/scaffold_test.go`

## 6. Update Self-References in Command Files

Each renamed command file may contain self-references
(its own name in headings, usage examples, error
messages). Update the content of all 10 embedded
assets AND their live copies.

- [x] 6.1 [P] Update self-references in
  `uf.address-feedback.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.address-feedback.md`, `.opencode/commands/uf.address-feedback.md`

- [x] 6.2 [P] Update self-references in
  `uf.agent-brief.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.agent-brief.md`, `.opencode/commands/uf.agent-brief.md`

- [x] 6.3 [P] Update self-references in
  `uf.cobalt-crush.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.cobalt-crush.md`, `.opencode/commands/uf.cobalt-crush.md`

- [x] 6.4 [P] Update self-references in
  `uf.constitution-check.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.constitution-check.md`, `.opencode/commands/uf.constitution-check.md`

- [x] 6.5 [P] Update self-references in
  `uf.finale.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.finale.md`, `.opencode/commands/uf.finale.md`

- [x] 6.6 [P] Update self-references in
  `uf.review-council.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.review-council.md`, `.opencode/commands/uf.review-council.md`

- [x] 6.7 [P] Update self-references in
  `uf.review-pr.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.review-pr.md`, `.opencode/commands/uf.review-pr.md`

- [x] 6.8 [P] Update self-references in
  `uf.triage-issue.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.triage-issue.md`, `.opencode/commands/uf.triage-issue.md`

- [x] 6.9 [P] Update self-references in
  `uf.init.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.init.md`, `.opencode/commands/uf.init.md`

- [x] 6.10 [P] Update self-references in
  `uf.unleash.md` (both embedded and live)
  Files: `internal/scaffold/assets/opencode/commands/uf.unleash.md`, `.opencode/commands/uf.unleash.md`

## 7. Update Cross-Command References

Commands that reference other uf commands by old
names. Update both embedded assets and live copies.

- [x] 7.1 Update `/uf.unleash` references to
  `/uf.cobalt-crush`, `/uf.review-council`,
  `/uf.finale`
  Files: `internal/scaffold/assets/opencode/commands/uf.unleash.md`, `.opencode/commands/uf.unleash.md`

- [x] 7.2 [P] Update `/uf.finale` references to
  `/uf.review-council`
  Files: `internal/scaffold/assets/opencode/commands/uf.finale.md`, `.opencode/commands/uf.finale.md`

- [x] 7.3 [P] Update `/uf.address-feedback`
  references to `/uf.review-council`
  Files: `internal/scaffold/assets/opencode/commands/uf.address-feedback.md`, `.opencode/commands/uf.address-feedback.md`

- [x] 7.4 [P] Update `/uf.init` references to
  `/uf.unleash`, `/uf.cobalt-crush`
  Files: `internal/scaffold/assets/opencode/commands/uf.init.md`, `.opencode/commands/uf.init.md`

- [x] 7.5 [P] Update `/uf.agent-brief` references
  to other uf commands (embedded AGENTS.md template)
  Files: `internal/scaffold/assets/opencode/commands/uf.agent-brief.md`, `.opencode/commands/uf.agent-brief.md`

## 8. Update Out-of-Scope-Owned Files

Files not owned by uf but containing references to
uf commands. Update to prevent stale handoffs.

- [x] 8.1 [P] Update `/opsx-propose` references to
  `/uf.unleash`, `/uf.cobalt-crush`
  Files: `.opencode/commands/opsx-propose.md`

- [x] 8.2 [P] Update `/opsx-explore` references to
  `/uf.unleash`
  Files: `.opencode/commands/opsx-explore.md`

- [x] 8.3 [P] Update `/opsx-apply` references to
  uf commands
  Files: `.opencode/commands/opsx-apply.md`

- [x] 8.4 [P] Update `speckit.implement` and other
  speckit commands referencing uf commands
  Files: `.opencode/commands/speckit.implement.md`
  and other `speckit.*.md` files as needed

- [x] 8.5 [P] Update agent files referencing uf
  commands: `cobalt-crush-dev.md`
  Files: `.opencode/agents/cobalt-crush-dev.md`

- [x] 8.6 [P] Update skill files referencing uf
  commands
  Files: `.opencode/skills/*/SKILL.md`

- [x] 8.7 [P] Update convention packs referencing
  uf commands
  Files: `.opencode/uf/packs/severity.md`

## 9. Update Documentation

- [x] 9.1 [P] Update `AGENTS.md`: PR Review
  Commands table, Issue Triage Commands table,
  Behavioral Rules section
  Files: `AGENTS.md`

- [x] 9.2 [P] Update `QUICKSTART.md`
  Files: `QUICKSTART.md`

- [x] 9.3 [P] Update `README.md`: Specification
  Framework section command references
  Files: `README.md`

- [x] 9.4 [P] Update `docs/usage.md`: command
  tables, usage examples
  Files: `docs/usage.md`

- [x] 9.5 [P] Update `docs/architecture.md`: file
  tree, command tables, review workflow
  Files: `docs/architecture.md`

- [x] 9.6 [P] Update `docs/heroes.md` if it
  references uf commands
  Files: `docs/heroes.md`

- [x] 9.7 [P] Update `CLAUDE.md` if it references
  uf commands
  Files: `CLAUDE.md`

## 10. Update Schema Samples

- [x] 10.1 [P] Update `gaze-hero.json`:
  `"/review-council"` → `"/uf.review-council"`
  Files: `schemas/hero-manifest/samples/gaze-hero.json`

- [x] 10.2 [P] Update `feedback-triage/README.md`
  and `issue-triage/README.md` if they reference
  old command names
  Files: `schemas/feedback-triage/README.md`, `schemas/issue-triage/README.md`

## 11. CHANGELOG and Migration Docs

- [x] 11.1 Add CHANGELOG entry under Unreleased
  section documenting the rename with migration
  reference to issue #302
  Files: `CHANGELOG.md`

## 12. Verification

- [x] 12.1 Run `make build` — verify the binary
  compiles with renamed embedded assets

- [x] 12.2 Run `make test` — verify all tests pass
  with updated assertions

- [x] 12.3 Run `make lint` — verify no lint issues

- [x] 12.4 Grep for old command names across the
  codebase. Verify no active references remain
  (exemptions: historical CHANGELOG entries,
  archived openspec changes, this proposal's own
  artifacts)

- [x] 12.5 Verify constitution alignment: no
  principle violated by the rename (per proposal
  assessment)

<!-- spec-review: passed -->
<!-- code-review: passed -->
