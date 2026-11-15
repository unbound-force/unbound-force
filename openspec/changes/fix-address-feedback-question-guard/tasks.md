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

## 1. Tighten command wording

Target: `.opencode/commands/uf.address-feedback.md` (canonical)

- [x] 1.1 In Section 3.2 Author Decision, change the per-item instruction at step 2 from "Present the **question tool** …" to an explicit imperative "Call the `question` tool … — do not write the call as text (`<invoke name=\"question\">`) in your reply" while retaining the visibility directive ("the printed text and the question tool MUST present the same details and options") and the recommended-first presentation order. (spec: address-feedback.md — Per-item triage decision presentation)
- [x] 1.2 Apply the same imperative to the Decision-table follow-up prompts (Modify / Reject / Ask) and to the triage-summary confirmation in Section 3.4 and the Phase 4 mandatory gates where a question tool call is required, using the same "Call … — do not write as text" phrasing. (spec)
- [x] 1.3 Sync the updated file byte-identically to `internal/scaffold/assets/opencode/commands/uf.address-feedback.md`. (FR mirror / drift gate)

## 2. Verification

- [x] 2.1 Run `go vet ./...` and `go test -race -count=1 ./internal/scaffold/...` (or `make check` subset) to verify scaffold drift — both copies must be byte-identical; confirm no test regressions in affected package.
- [x] 2.2 Verify constitution alignment (Principle VI PASS per proposal): confirm no new deterministic procedure is left in prose that belongs in a tool — the in-repo change is wording-only; the deterministic idle-recovery procedure remains in `opencode-supervisor#65` / `gh-hud#9` out-of-scope per design D1.
