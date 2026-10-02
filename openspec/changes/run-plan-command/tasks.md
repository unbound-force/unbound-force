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

## 1. Command Authoring

- [x] 1.1 Create canonical command file `.opencode/commands/uf.run-pipeline-plan.md` with full command definition including:
  - Frontmatter (description, usage)
  - Input detection logic (file path vs. issue refs)
  - Phase 0: input resolution, `gh issue view` calls, work-item table confirmation via `question`, non-mechanical item flagging
  - Per-item protocol: fresh `general` subagent per phase (`/opsx-propose` -> `/uf.unleash` -> `/uf.finale`), disk handoff verification, clean tree confirmation between items
  - Gate relay protocol: subagents return `GATE:<name>` + question + text + options; orchestrator relays via `question` and resumes same subagent via `task_id`; subagents forbidden from invoking `question` directly
  - Guardrails: scaffold drift sync, explicit staging (never `git add -A`/`git add .`), remove stray `internal/scaffold/AGENTS.md` before staging, no `git push --force` (only `--force-with-lease`), PR targets `unbound-force/unbound-force` head `jflowers:<branch>`, `Closes #N` in commit message and PR body, never merge PR
  - Tracking: `todowrite` one entry per work item, complete only after PR created and CI green
  - Plan file contract: numbered list with issue refs, `Change:` name, optional description, `[INVESTIGATION]`/`[POLICY]` tags for non-mechanical items

## 2. Scaffold Sync

- [x] 2.1 Copy canonical command file to embedded location: `cp .opencode/commands/uf.run-pipeline-plan.md internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md` (MUST be byte-identical)
- [x] 2.2 Add `"opencode/commands/uf.run-pipeline-plan.md"` to `expectedAssetPaths` in `internal/scaffold/scaffold_test.go` (after line 153, maintaining alphabetical order within the OpenCode commands group)

## 3. Verification

- [x] 3.1 Run `make test` to verify `TestEmbeddedAssets_MatchSource` passes (canonical and embedded copies are byte-identical) and `TestAssetPaths_MatchExpected` passes (manifest is current)
- [x] 3.2 Run `make lint` to verify no lint errors in modified files
- [x] 3.3 Verify constitution alignment: confirm all five principles (Autonomous Collaboration, Composability First, Observable Quality, Testability, Security by Default) remain PASS as documented in proposal.md

<!-- spec-review: passed -->
<!-- code-review: passed -->
