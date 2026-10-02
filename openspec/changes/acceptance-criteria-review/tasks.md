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

## 1. Originating-Issue Threading (review-context skill)

- [x] 1.1 Update `.opencode/skills/review-context/SKILL.md` Protocol 2 to add originating-issue resolution priority: (1) `originating_issue` field in `.openspec.yaml`, (2) PR body parsing fallback, (3) no originating issue. Add a step to read `.openspec.yaml` from the active change directory before parsing the PR body.
- [x] 1.2 Add Given/When/Then scenarios to Protocol 2 documentation covering: field present, field absent, field conflicts with PR body references.

## 2. Acceptance-Criteria Evaluation and Deviation Escape Hatch (review-pr command)

- [x] 2.1 Update `.opencode/commands/uf.review-pr.md` Step F.1 to add two evaluation modes: rigorous (SATISFIED / NOT SATISFIED / PARTIAL per Given/When/Then scenario with diff evidence) and best-effort (COVERED / NOT COVERED / PARTIAL for freeform criteria). Auto-detect mode from criteria structure.
- [x] 2.2 Add `IMPLEMENTATION_DEVIATION` finding type to Step F.1: severity HIGH, not auto-fixable, requires criterion reference + reason + governance action. Document that deviations with adequate documented governance result in COMMENT verdict (not REQUEST CHANGES) when no other REQUEST CHANGES findings exist.
- [x] 2.3 Update Step F.1 output format to include per-criterion evaluation results in the "Linked Issues" section: show SATISFIED/NOT SATISFIED/PARTIAL (or COVERED/NOT COVERED/PARTIAL) per criterion/scenario with evidence.

## 3. Review-Council Guard Persona Updates

- [x] 3.1 Update `.opencode/commands/uf.review-council.md` Step 2 (Divisor agent delegation) to instruct the Guard persona to: (a) receive originating-issue acceptance criteria when resolved, (b) evaluate whether unmet criteria have an `IMPLEMENTATION_DEVIATION` with adequate governance, (c) flag missing governance as a HIGH-severity finding.
- [x] 3.2 Update Phase 1c (Protocol 2 invocation) to note that the review-context skill now resolves originating issues from `.openspec.yaml` first, and to pass the resolved issue context to the Guard persona in Step 2.

## 4. Council-Review-Action Prompt Threading

- [x] 4.1 [P] Update `council-review-action/scripts/build-prompt.sh` to detect and pass `originating_issue` from `.openspec.yaml` into the review prompt context when available.
- [x] 4.2 [P] Update `council-review-action/scripts/run-review.sh` to include originating-issue acceptance criteria in the Divisor agent prompts when resolved.

## 5. Verification and Constitution Alignment

- [x] 5.1 Verify constitution alignment: confirm that all changes maintain PASS assessments for principles I (Autonomous Collaboration), II (Composability First), III (Observable Quality), and IV (Testability) as documented in the proposal.
- [x] 5.2 Test the originating-issue resolution priority: create a test fixture with `.openspec.yaml` containing `originating_issue: N` and verify Protocol 2 resolves to issue N without parsing the PR body.
- [x] 5.3 Test the deviation escape hatch: create a test fixture with an unmet criterion and a documented governance action, verify the reviewer emits `IMPLEMENTATION_DEVIATION` with severity HIGH and COMMENT verdict.
- [x] 5.4 Run `/uf.review-council` on the change branch to verify the spec artifacts pass spec review.

<!-- spec-review: passed -->
<!-- code-review: passed -->
