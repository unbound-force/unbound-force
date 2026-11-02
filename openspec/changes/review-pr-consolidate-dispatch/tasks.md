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

## 1. Rewrite `/uf.review-pr` finalization

All tasks in this group modify the same file
(`.opencode/commands/uf.review-pr.md` and its
`internal/scaffold/assets/opencode/commands/uf.review-pr.md`
mirror, kept byte-identical within each task) — no parallel
execution.

- [ ] 1.1 Replace Sections 5, 5a, and 5b with a single
      "Finalize Dispatch" section (both copies): exactly one
      `consolidate_dispatch` call per finalization attempt,
      passing the captured `correlation_id` from the dispatch
      step and `provenance` with `workflow_id`
      `"uf.review-pr"`; `branch` and `commit` resolve only
      from the Step 2 frozen PR input context (`head_ref`,
      `head_sha`), never from the local checkout, with a
      missing or invalid frozen field failing closed with an
      explicit error naming the field; state plainly "Never
      assemble, restate, or recompute the payload; never call
      `finalize_review_dispatch` directly"; read the
      review-dispatch artifact path, canonical review-verdict
      v2 projection, and calculated-assessment authority
      directly from the tool response; handle the all-skip
      dispatch (no session exists; non-mechanical;
      fail-closed path applies); bounded mechanical retry of
      the identical call at most twice, classified from the
      structured response alone (default-deny), with
      exhausted retries or non-mechanical failures failing
      closed to `INCONCLUSIVE` (FR-013; FR-002, FR-003,
      FR-004 applied unchanged); in the same rewrite, update
      the Section 4c dispatch-capture prose (~L446-460) that
      cross-references the deleted Section 5a/5b paths and
      the no-provenance retrieval, so Section 4c forwards the
      captured `correlation_id` into the single finalization
      call and sources proposals and warnings from its result
      (FR-013, FR-014)
- [ ] 1.2 In the same rewrite, scope deduplication,
      cross-persona root-cause grouping, and compound
      severity to presenting the review comment and terminal
      report only, stating explicitly that they never merge,
      filter, re-order, or otherwise alter findings,
      contributing run ids, verdicts, run counts, or
      advisories in finalization data; remove the
      finalization-path verdict-ladder prose — the verdict
      and zero-success results come from the tool's
      finalization result (FR-015)
- [ ] 1.3 In the same rewrite, source lesson proposals and
      consolidation warnings from the finalization result's
      top-level `proposals` and `warnings` arrays; remove the
      Section 5a no-provenance retrieval call and its
      retrieval-failure behavior block; keep the
      `dewey_store_learning` contract (exactly once per
      proposal, using the proposal's `information`, `tag`,
      and `category`; never raw child `> learn:` text or
      child-supplied tags, categories, or hashes); on the
      fail-closed path, report stranded proposals as
      retained in the dispatch store and store nothing
      fabricated (FR-014)

## 2. Tidy `/uf.review-council` Section 3

- [ ] 2.1 Reword the Section 3 (~L360) "First deduplicate
      ... cross-persona grouping" prose in
      `.opencode/commands/uf.review-council.md` and its
      `internal/scaffold/assets/opencode/commands/`
      `uf.review-council.md` mirror to state explicitly that
      deduplication and grouping govern presentation of the
      review comment and terminal report only and never
      finalization data; leave Section 5 (already
      consolidate-with-provenance) unchanged (FR-015)

## 3. Content-gate tests and site-label refresh

- [ ] 3.1 Add content-gate assertions in
      `internal/scaffold/scaffold_test.go` (following the
      `TestQuestionToolPresentationContractGates` precedent)
      reading the canonical command sources: for
      `uf.review-pr.md`, assert presence of the one-call
      `consolidate_dispatch` finalization contract with
      `workflow_id` `"uf.review-pr"`, the frozen-context
      provenance source (`head_ref`/`head_sha`), the
      never-assemble and never-call-finalizer-directly
      prohibitions, the bounded retry policy, and the
      presentation-only grouping scope; for
      `uf.review-council.md`, assert the presentation-only
      scope of the Section 3 deduplication prose (FR-016)
- [ ] 3.2 Add absence assertions to the same content-gate
      test: `uf.review-pr.md` no longer contains the legacy
      manual-finalize fragments — pin finalization-unique
      removed strings (the Section 5b "Call
      `finalize_review_dispatch` with" instruction and the
      finalization-path verdict-ladder sentence "Any blocking
      successful run yields `REQUEST CHANGES`", which task
      1.2 removes outright), and scope the
      deduplication-mandate absence to the Finalize Dispatch
      section boundaries rather than whole-file matching of
      "First deduplicate" (which legitimately remains in
      presentation-scoped prose), so the prohibitions cannot
      be silently reintroduced (FR-016)
- [ ] 3.3 [P] Refresh the documented-site inventory in
      `.opencode/test/golden-output-shape.test.ts` (design
      D6): the coverage comments and the `it.each` site
      labels no longer attribute a `finalize_review_dispatch`
      site to `uf.review-pr.md` (keep the underlying
      plugin-tool coverage via the remaining documented site,
      `speckit.testreview.md:278`); in the same refresh,
      document the new `uf.review-pr.md` site of the
      finalization `consolidate_dispatch` call in the
      coverage comments and `it.each` inventory and update
      the per-plugin site counts so the inventory stays
      exhaustive
- [ ] 3.4 [P] Update the `uf.review-pr.md` clause set pinned
      by `TestCommandContracts_CorrelationIdForwarding` in
      `internal/scaffold/command_contract_test.go`
      (~L302-381): its current clauses assert the deleted
      Section 5a/5b strings over both the live command and
      its scaffold mirror; re-pin them to the new one-call
      contract — correlation-id forwarding into the single
      `consolidate_dispatch` finalization call and its
      bounded identical retries (FR-013, FR-016)

## 4. Verification, Parity, and Docs

- [ ] 4.1 Verify byte-identical parity between each edited
      `.opencode/commands/` file and its
      `internal/scaffold/assets/opencode/commands/` mirror
      (`uf.review-pr.md`, `uf.review-council.md`) (FR-016,
      FR-005)
- [ ] 4.2 Run `make check` — which includes the scaffold
      drift-detection and content-gate tests, `make build`
      (embed.FS asset sync), and `make plugin-test`; all must
      pass (FR-016)
- [ ] 4.3 Add a CHANGELOG.md entry describing the
      deterministic one-call finalization for `/uf.review-pr`
      and the presentation-only grouping scope
- [ ] 4.4 File both documentation issues for the user-facing
      command behavior change before PR merge: the repo-local
      documentation issue (per the AGENTS.md documentation
      gate; duplicate-check open docs issues in this repo
      first) and the `unbound-force/website` tracking issue
      (per the constitution's Cross-Repo Documentation rule;
      duplicate-check open website issues first — #303, #316,
      and #360 do not cover this change)
- [ ] 4.5 Confirm constitution alignment: human-confirmation
      gates untouched, no quality or governance gate
      modified, finalization artifacts remain
      tool-authoritative (proposal Constitution Alignment
      I-VI)
