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

## 1. Pre-flight Skill: Add `soft-gate` Mode

All tasks in this group modify a single file
(`.opencode/skills/pre-flight/SKILL.md`), so no tasks
are parallel-eligible.

- [x] 1.1 Update the Execution Policies table in SKILL.md
  to add the `soft-gate` row:
  `| soft-gate | Run all tools. Classify failures as
  branch-caused vs pre-existing. Gate only on
  branch-caused. | /review-council |`
  Update the description line to mention all three
  policies.
  **File**: `.opencode/skills/pre-flight/SKILL.md`
  **Spec**: FR-001

- [x] 1.2 Add Phase 4 `soft-gate` mode section to
  SKILL.md. After the existing `hard-gate mode` and
  `ci-aware mode` subsections in Phase 4, add a
  `soft-gate mode` subsection that describes:
  (a) Execute all tools without stopping on failure.
  Record all exit codes and output.
  (b) If all tools pass, verdict is PASS (no baseline
  needed).
  (c) If any tools fail, proceed to baseline
  establishment (new Phase 4a).
  **File**: `.opencode/skills/pre-flight/SKILL.md`
  **Spec**: FR-001, D5

- [x] 1.3 Add Phase 4a: Baseline Establishment section
  to SKILL.md (new section between Phase 4 and Phase 5).
  Document the two-tier baseline strategy:
  **Tier 1 -- CI API**: Check `which gh`. If available,
  query `gh api repos/{owner}/{repo}/commits/main/check-runs
  --jq '.check_runs[] | {name, conclusion}'`. Use
  `--arg` for dynamic values. Map CI check names to
  local tool names via coverage matrix.
  **Tier 2 -- Worktree**: If CI API unavailable or no
  data, create `git worktree add
  /tmp/preflight-baseline-<SHORT_SHA> main --detach`.
  Run ONLY failing tools in worktree. Clean up with
  `git worktree remove ... --force`.
  **Fallback**: If both tiers fail, classify all
  failures as `unknown` (treated as branch-caused).
  **File**: `.opencode/skills/pre-flight/SKILL.md`
  **Spec**: FR-002, FR-003, D2

- [x] 1.4 Add Phase 4b: Causality Classification section
  to SKILL.md (after Phase 4a). Document the
  classification table:
  | Baseline status | Branch status | Classification |
  | Pass | Fail | branch-caused |
  | Fail | Fail | pre-existing |
  | No data | Fail | unknown (treat as branch-caused) |
  After classification, apply gate decision:
  branch-caused or unknown → FAIL verdict.
  pre-existing only → PASS verdict.
  **File**: `.opencode/skills/pre-flight/SKILL.md`
  **Spec**: FR-001, D3, D4

- [x] 1.5 Update Phase 5 result format in SKILL.md to
  add `soft-gate` output. Extend the Execution Results
  table with a `Causality` column. Extend the Verdict
  block with `Branch-caused failures`,
  `Pre-existing failures`, and `Baseline method` fields.
  Document that Result is `PASS` when only pre-existing
  failures exist, and `FAIL (branch-caused)` when any
  branch-caused or unknown failures exist.
  **File**: `.opencode/skills/pre-flight/SKILL.md`
  **Spec**: FR-004, D6

## 2. Review-council: Switch to `soft-gate`

All tasks in this group modify a single file
(`.opencode/commands/review-council.md`), so no tasks
are parallel-eligible.

- [x] 2.1 Update Phase 1a in review-council.md to use
  `soft-gate` instead of `hard-gate`. Change the
  heading from "Pre-flight Checks (mandatory, hard gate)"
  to "Pre-flight Checks (mandatory, soft gate)".
  Update the instruction text to: "Load the `pre-flight`
  skill and run in `soft-gate` mode."
  Update the gate logic:
  (a) If verdict is FAIL (branch-caused): STOP
  immediately. Report branch-caused failures as
  CRITICAL findings. Do NOT proceed to Phase 1b.
  (b) If verdict is PASS (including when pre-existing
  failures exist): report success, record any
  pre-existing failures for Step 6, proceed to
  Phase 1b.
  **File**: `.opencode/commands/review-council.md`
  **Spec**: Modified -- Phase 1a

- [x] 2.2 Update Step 6 (final report) in
  review-council.md to include a "Pre-existing CI
  Failures (informational)" section. Add this between
  the discovery summary and the iteration findings.
  Include a table with Tool, Exit code, and Baseline
  method columns. Add note: "These do not block the
  review verdict." Omit the section when no
  pre-existing failures were detected.
  **File**: `.opencode/commands/review-council.md`
  **Spec**: Modified -- Step 6, D7

## 3. Scaffold Sync

These tasks modify different files from each other
and from groups 1-2, so they are parallel-eligible.

- [x] 3.1 [P] Copy the updated
  `.opencode/skills/pre-flight/SKILL.md` to
  `internal/scaffold/assets/opencode/skills/pre-flight/SKILL.md`.
  The two files MUST be byte-identical. The drift
  detection test `TestEmbeddedAssets_MatchSource`
  enforces this at build time.
  **File**: `internal/scaffold/assets/opencode/skills/pre-flight/SKILL.md`

- [x] 3.2 [P] Copy the updated
  `.opencode/commands/review-council.md` to
  `internal/scaffold/assets/opencode/commands/review-council.md`.
  The two files MUST be byte-identical. The drift
  detection test `TestEmbeddedAssets_MatchSource`
  enforces this at build time.
  **File**: `internal/scaffold/assets/opencode/commands/review-council.md`

## 4. Verification

- [x] 4.1 Run `make check` (build, lint, test) to
  verify no regressions. The
  `TestEmbeddedAssets_MatchSource` test MUST pass,
  confirming scaffold sync.

- [x] 4.2 Constitution alignment verification: confirm
  the implementation maintains alignment with all five
  org constitution principles as assessed in the
  proposal:
  - I. Autonomous Collaboration: skill remains a
    self-contained artifact, no runtime coupling added
  - II. Composability First: `hard-gate` and `ci-aware`
    unchanged, no mandatory dependencies introduced
  - III. Observable Quality: result format includes
    machine-parseable causality classification
  - IV. Testability: classification logic is
    deterministic (exit code comparison)
  - V. Security by Default: `--arg` used for injection
    safety, worktree cleanup, no new dependencies

- [x] 4.3 Manual verification: test the three key
  scenarios from the spec:
  (a) Branch-caused failure: introduce a lint error on
  the branch (not on `main`), run `/review-council`,
  confirm it stops with CRITICAL finding.
  (b) Pre-existing failure: if `main` has a failing
  check, create a branch, run `/review-council`,
  confirm it proceeds and reports the failure as
  informational.
  (c) All pass: on a clean branch, run
  `/review-council`, confirm PASS verdict and normal
  council flow.
<!-- spec-review: passed -->
<!-- code-review: passed -->
