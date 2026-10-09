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

## 1. AGENTS.md — Commit Scope and Documentation Gate

- [x] 1.1 Update the **Commit scope** behavioral rule in `AGENTS.md` (line ~192) to add explicit clarification: "Dewey learnings (`.uf/dewey/learnings/*.md` and `.uf/dewey/compiled/*.md`) produced during the change's workflow are directly related to the active change and MUST be included in the feature PR."
- [x] 1.2 Add a note to the **Documentation gate** section in `AGENTS.md` (line ~181) stating that retrospective learnings satisfy the documentation gate as intentional knowledge-capture artifacts.

## 2. uf.unleash.md — Step 9 Policy Annotation

- [x] 2.1 Add a `> POLICY NOTE:` blockquote to Step 9 (Retrospective) in `.opencode/commands/uf.unleash.md` (after line ~670) stating that learnings stored via `dewey_store_learning` produce files (`.uf/dewey/learnings/*.md`, `.uf/dewey/compiled/*.md`) that are part of the feature PR scope and MUST be committed on the feature branch.

## 3. Verification

- [x] 3.1 Verify constitution alignment: confirm that the updated Commit scope rule and Documentation gate note are consistent with Observable Quality (traceability/provenance) and Autonomous Collaboration (artifact-based communication) as assessed in the proposal.
- [x] 3.2 Run `make lint` and `make build` to verify no structural breakage from markdown changes (documentation-only change; these commands validate repo health).
<!-- spec-review: passed -->
<!-- code-review: passed -->
<!-- scaffolded by uf vdev -->
