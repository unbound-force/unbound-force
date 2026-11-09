## Context

`projectSummaryProposal` (`.opencode/plugins/review-dispatch/index.ts`)
projects each proposal for the summary retrieval
(`buildSummaryRetrieval`) and the compact finalization result
(`buildCompactFinalization`). It cut `information` to 512
characters and `agent` / `tag` to 128, marking the entry
`truncated: true`. The persisted artifact (RD-FR-054) and the full
retrieval keep complete text. Issue #969 shows a parent storing the
cut summary copy despite the RD-FR-056 prose contract.

## Goals / Non-Goals

### Goals

- No `consolidate_dispatch` result can hand a parent a shortened
  proposal.
- Keep every tool result under the 40,960-byte budget.
- Keep findings and advisories display caps unchanged.

### Non-Goals

- No change to `dewey_store_learning` (dewey repo).
- No heuristic completeness check on lesson bodies (sentence
  terminator detection): it would reject legitimate lessons that
  end in an identifier or code span, and as command prose it would
  violate DE-001.
- No change to the store-from-artifact contract (RD-FR-056).

## Decisions

### D1. Proposals are whole or absent

A previewed proposal copies `agent`, `information`, `tag`, and a
valid `category` verbatim. Proposals reaching the projection already
passed `SubmitLessonProposalInputSchema` at harvest, and `agent` is
the run's validated agent name, so fields are bounded (16,384 /
128). The per-entry `truncated` marker is never set
on a proposal; the only damage signal is `proposals_truncated`,
meaning whole entries were dropped. A consumer therefore cannot
persist a partial body from any retrieval mode.

Rejected: keeping the 512 cut with a louder marker. The marker
already existed (`truncated: true`) and was ignored in PR #955;
the safe fix removes the damaged data rather than labeling it.

### D2. Byte budget on the compact finalization result

`buildCompactFinalization` previously relied on per-field caps for
its bound. With full-length proposals, ten 16,384-character entries
reach about 160 KiB. The builder now assembles the compact result,
measures its compact UTF-8 serialization, and drops trailing
proposals until it is strictly under `SUMMARY_LIMITS.maxBytes`,
setting `proposals_truncated`. Key order is unchanged.

The summary builder keeps its existing drop order (findings,
advisories, proposals, warnings): proposals outrank findings in the
summary because findings are re-read from the artifact.

### D3. Prose corrections only

`uf.review-council.md` Section 4 and `uf.review-pr.md` Section 5
described the preview as "each `information` cut to 512
characters". The sentence now reads "at most 10 complete entries;
`proposals_truncated` marks dropped ones". Lesson storage still
reads `payload.proposals` through `read_dispatch_artifact`.
`uf.address-feedback.md` and `uf.triage-issue.md` carry no 512
claim and are not edited.

## Risks / Trade-offs

- **Summary findings displaced by long proposals.** Large lessons
  can push findings out of the summary preview. Findings are
  authoritative only in the artifact, and the drop is flagged by
  `findings_truncated`.
- **Coverage floor.** The new budget loop is covered by unit tests.
  If the 94 percent floor fails, the Gatekeeping rule applies.

## Archive order

This change amends clauses of the unarchived
`bound-consolidate-dispatch-output` (RD-FR-049, RD-FR-051) and
`review-dispatch-consumer-bounds` (RD-FR-054) changes. Archive it in
the same batch as, or after, both.
