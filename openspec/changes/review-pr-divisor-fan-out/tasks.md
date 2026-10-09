<!--
  [P] marks tasks eligible for parallel execution.
  All tasks below modify the same file
  (.opencode/commands/uf.review-pr.md), so none are
  marked [P] — parallel execution would cause merge
  conflicts.
-->

## 1. Study Reference Files

- [ ] 1.1 Read `.opencode/commands/uf.review-council.md` (the target
  fan-out pattern to mirror)
- [ ] 1.2 Read `.opencode/commands/uf.review-pr.md` (the current
  command to rewrite)
- [ ] 1.3 Read `.opencode/skills/review-context/SKILL.md` (Protocols
  1-4 used in Step 3.8)
- [ ] 1.4 Read `.opencode/skills/dispatch-advisor/SKILL.md` (fan-out
  pattern used in Steps 4a-6)

## 2. Rewrite Parent Pipeline Steps (0 through 3.10)

- [ ] 2.1 Preserve Steps 0-3 (prerequisites, resolve PR, fetch
  metadata, CI checks + causality) — these do not change
- [ ] 2.2 Preserve Step 3.5 (diff size check) — becomes advisory
  only, no longer blocks delegation
- [ ] 2.3 Add Step 3.6: Fetch PR diff using
  `replicator_forge_adversarial_review` or `gh pr diff` — moved from
  old subagent Step B
- [ ] 2.4 Add Step 3.7: Run pre-flight skill in `ci-aware` mode —
  moved from old subagent Step A
- [ ] 2.5 Add Step 3.8: Run review-context skill (Protocols 1-4,
  including issue linking from diff) — moved from old subagent Step C
- [ ] 2.6 Add Step 3.9: Load convention packs (default, go, typescript,
  severity, content) — moved from old subagent Step D
- [ ] 2.7 Add Step 3.10: Fetch existing GitHub review state (prior
  reviews, CODEOWNER status) — moved from old subagent Step E

## 3. Add Multi-Agent Divisor Fan-Out (Steps 4 through 6)

- [ ] 3.1 Add Step 4: Discover available divisor-* agents by listing
  `.opencode/agents/divisor-*.md` files
- [ ] 3.2 Add Step 4a: Load dispatch-advisor skill
- [ ] 3.3 Add Step 4b: Call `plan_review_dispatch` with mode `code`,
  `changed_files` from PR metadata, and `discovered_agents`
- [ ] 3.4 Add Step 4c: Call `acquire_sibling_evidence`
- [ ] 3.5 Add Step 5: For each planned run, invoke `invoke_agent` with
  agent name, model, variant from the dispatch plan, and a prompt
  containing diff + review context + convention packs + CI results +
  PR metadata
- [ ] 3.6 Add Step 6: Consolidate findings from all agent runs, prepare
  Dewey lessons, and call `finalize_review_dispatch` with command
  `"review-council"` to produce the structured verdict

## 4. Rewrite Output and Post-Review Steps (7 through 9)

- [ ] 4.1 Rewrite Step 7: Adapt output format for council results —
  show per-agent findings, consolidated verdict, and advisories
- [ ] 4.2 Preserve Step 8: Fix-branch offer (unchanged from current
  command)
- [ ] 4.3 Preserve Step 9: Verdict-aligned PR review posting using
  `gh pr review` — map council verdict to APPROVE / REQUEST_CHANGES /
  COMMENT

## 5. Remove Old Subagent Block

- [ ] 5.1 Remove Steps A-F (BEGIN SUBAGENT PROMPT through END
  SUBAGENT PROMPT) — these are replaced by Steps 4-6

## 6. Verify

- [ ] 6.1 Verify the new command mirrors review-council's fan-out
  pattern (discover → dispatch-advisor → plan_review_dispatch →
  invoke_agent → consolidate → finalize)
- [ ] 6.2 Verify constitution alignment: Autonomous Collaboration
  (structured artifact output), Composability First (independent agent
  invocation), Observable Quality (finalize_review_dispatch provenance)
- [ ] 6.3 Verify no Go code, CI config, or test files need changes
  (command markdown only)