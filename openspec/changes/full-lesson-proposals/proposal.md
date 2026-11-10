## Why

`consolidate_dispatch` previews lesson proposals in two bounded tool
results: the `detail: "summary"` retrieval (RD-FR-049) and the
compact finalization result (RD-FR-051). Both cut each proposal's
`information` to 512 characters (`SUMMARY_LIMITS.informationChars`)
and mark the entry `truncated: true`. RD-FR-056
(`review-dispatch-consumer-bounds`) made the persisted artifact the
only lesson source and told every command never to store from the
preview, but that contract lives in prose. In PR #955 the parent
stored from the summary anyway:

- 5 of 6 council iteration-1 learnings (commit 64457c88) ended
  mid-word ("...after i", "...then subm", "...the push-fa"); the
  sixth lost its final sentence. All 6 iteration-2 learnings were
  cut the same way.
- All 30 feedback-assessment learnings of the same PR are
  complete: they came from artifact reads, which never cut text.

A 512-character cap is a display bound for findings. Proposals are
records a parent persists verbatim, and a body cut mid-word loses
the conclusion of the gotcha it records. A command-side guard
("reject a body without a sentence terminator") would be mechanical
prose (DE-001) and a heuristic; removing the cut at the source makes
every retrieval mode safe to harvest.

Fixes #969.

## What Changes

- The summary retrieval and the compact finalization result copy
  every proposal field verbatim. `SUMMARY_LIMITS` drops
  `informationChars`, `tagChars`, and `agentChars`; proposal fields
  stay bounded by `PayloadLessonProposalSchema` (`information`
  16,384, `tag` 128) at harvest time.
- Completeness guarantee: a previewed proposal is either complete
  or absent. Entries that do not fit the 40,960-byte budget are
  dropped whole (trailing first) and counted by
  `proposals_truncated: true`. A proposal entry never carries
  `truncated`.
- The compact finalization result gains the same post-assembly
  byte budget for its proposals, since entries are no longer
  bounded per field.
- Findings, advisories, warnings, and `verdict_reason` keep their
  RD-FR-049 caps and per-entry `truncated` markers.
- `uf.review-council.md` Section 4 and `uf.review-pr.md` Section 5
  describe the preview as "at most 10 complete entries" instead of
  "cut to 512 characters". The store-from-artifact contract
  (RD-FR-056) is unchanged.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: complete lesson proposals in every tool result
  (RD-FR-058, the id `review-dispatch-consumer-bounds` left
  reserved). Amends the proposal clauses of RD-FR-049, RD-FR-051,
  and RD-FR-054.

### Removed Capabilities
- None.

## Impact

- **Live plugin**: `.opencode/plugins/review-dispatch/index.ts` —
  `SUMMARY_LIMITS`, `projectSummaryProposal`,
  `buildCompactFinalization`, and the tool description.
- **Commands**: one sentence each in `uf.review-council.md`
  Section 4 and `uf.review-pr.md` Section 5.
  `uf.address-feedback.md` is not edited.
- **Scaffold mirror**: every edited `.opencode/` file is copied
  byte-identical under `internal/scaffold/assets/opencode/`.
- **Tests**: vitest coverage in `summary-retrieval.test.ts` and an
  end-to-end harvest test in `consolidate-dispatch.test.ts`
  (`submit_lesson_proposal` -> `dispatch_review_runs` ->
  summary, full, finalization, `read_dispatch_artifact`).
- **Out of scope**: `dewey_store_learning` lives in the dewey repo;
  no change there.
- **CI replication**: `make check`.
- **Documentation impact**: CHANGELOG.md entry. The edited command
  prose is user-facing — `uf init` deploys the mirrored commands —
  so the website documentation surface applies:
  unbound-force/website#406 tracks it (website#303 precedent), and
  in-repo docs issue #1036 covers the documentation impact.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

Every tool result now carries lessons a consumer can store as-is,
so a caller that misreads the preview contract still persists the
complete record.

### II. Composability First

**Assessment**: PASS

The change stays inside the review-dispatch plugin. Result field
names and key order are unchanged; only shortening is removed.

### III. Observable Quality

**Assessment**: PASS

Damage stays visible: a proposal that does not fit is dropped whole
and flagged by `proposals_truncated`, never silently shortened.

### IV. Testability

**Assessment**: PASS

The projection and budget loops are pure functions with unit tests;
the harvest path is tested end to end with the real submit tool and
an injected fake client.

### V. Security by Default

**Assessment**: PASS

No new input surface. Proposals were Zod-validated at submit and
harvest time; the byte budget still bounds every tool result.

### VI. Deterministic Execution

**Assessment**: PASS

The completeness guarantee is enforced in tool code. No mechanical
prose is added to commands; two sentences are corrected.
