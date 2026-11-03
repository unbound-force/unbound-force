<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the same group, (b) has no dependency
  on prior tasks in the group, and (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Plugin contract (design D1-D3)

- [x] 1.1 In `.opencode/plugins/review-dispatch/index.ts`, add the
      strict `PayloadLessonProposalSchema` (`agent`, `information` <= 16,384, `tag` <= 128, optional `category`) and an optional `proposals` array (max 512) to
      `ReviewDispatchPayloadSchema`; order `proposals` after `advisories` and before `plan` in `artifactPayloadOrder` (RD-FR-054).
- [x] 1.2 Add `deduplicateProposals` (NFKC, lowercase, letter/digit
      tokens, fixed stop words, overlap coefficient >= 0.5 for token sets of 4 or more, exact token sequence otherwise, longer text wins, earlier slot kept) and the
      `duplicate_proposal_merged` warning with its `agent=<agent> kept=<kept_agent>` summary detail (RD-FR-055).
- [x] 1.3 Feed the deduplicated proposals into the finalizer payload
      (omitted when empty), the full and summary retrievals, and the compact result; update the tool description to call the tool-result arrays a preview and the
      artifact the lesson source (RD-FR-054, D3).
- [x] 1.4 Add `proposals` and `$defs/lessonProposal` to
      `schemas/review-dispatch/v1.0.0.schema.json`, with a Go schema test in `internal/schemas/` covering a valid and a malformed proposal.
- [x] 1.5 Vitest coverage: persisted full text, artifact key order,
      empty omission, strict rejection of malformed proposals, the #941 duplicate pair, distinct lessons kept, short-proposal exact matching, and warning projection.

## 2. Consumer commands (design D4-D5)

- [x] 2.1 `uf.review-pr.md` Section 5: store lessons from the
      artifact's `payload.proposals`; the compact arrays are a preview.
- [x] 2.2 `uf.review-council.md` Sections 2, 4, and 5: keep the
      summary retrieval for presentation and warnings; store lessons after Section 5 from that iteration's artifact; add `duplicate_proposal_merged` to the warning kinds.
- [x] 2.3 `uf.unleash.md` Steps 4 and 6: store lessons from the
      artifact each step already reads.
- [x] 2.4 `uf.triage-issue.md` Sections 2.3, 2.4, 3.1, 3.2, and 4.6:
      finalize immediately after dispatch, read runs, findings, advisories, and proposals from the artifact, and drop the no-provenance full retrieval.
- [x] 2.5 `uf.address-feedback.md` Sections 2.3.4, 2.3.5, 2.3.7, and
      2.3.8: finalize each item immediately after its dispatch and read its artifact; drop the no-provenance full retrieval.
- [x] 2.6 Update the Go content gates in `internal/scaffold/` for the
      new wording, including gates that pin the removed retrieval prose.

## 3. Spec deltas for `diff_not_found` (design D6)

- [x] 3.1 [P] `openspec/changes/build-review-prompts/specs/build-review-prompts.md`:
      FR-011 in-scope missing-path rule and scenarios; FR-009 not-found scenario becomes not-readable (#943).
- [x] 3.2 [P] `openspec/changes/review-dispatch-diff-manifest/specs/review-dispatch/spec.md`:
      `read_diff_chunk` `diff_not_found` scenario (#943).

## 5. Review follow-ups (design D2, D7)

- [x] 5.1 Copy the updated `schemas/review-dispatch/v1.0.0.schema.json`
      to `.uf/schemas/` and `internal/scaffold/assets/uf/schemas/`; add `TestSchemaRegistryMatchesUFCopies`.
- [x] 5.2 Switch deduplication to Jaccard >= 0.3 with a 6-token floor
      and a negation guard; test the negated and subset pairs.
- [x] 5.3 Cap persisted proposals at 512 with a `proposals_overflow`
      warning; test 515 proposals.
- [x] 5.4 Add `read_dispatch_artifact` (RD-FR-057) with tests, and move
      every consumer's artifact read onto it, including the older-plugin stranded rule.
- [x] 5.5 Nits: Section 3.2 reference, 512-character bound including
      the ellipsis, pre-dedupe proposal count note.

## 4. Mirror, documentation, CI parity

- [x] 4.1 Copy every edited `.opencode/` file byte-identical into
      `internal/scaffold/assets/opencode/`.
- [x] 4.2 CHANGELOG.md entries under Unreleased with
      `(Spec: openspec/changes/review-dispatch-consumer-bounds, Fixes: #N)` tails.
- [ ] 4.3 Documentation-gate assessment; file the docs issue for the
      user-facing artifact and consumer change before merge.
- [x] 4.4 CI parity: `make plugin-test` and
      `go test -race -count=1 ./...` both pass.
