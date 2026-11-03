## Context

`createConsolidateDispatchTool`
(`.opencode/plugins/review-dispatch/index.ts`) collects lesson
proposals from successful runs into `collectedProposals` and returns
them only on the tool result: capped and cut in the summary
retrieval (RD-FR-049) and in the compact finalization result
(RD-FR-051). RD-FR-050 kept them out of the payload handed to
`finalizeReviewDispatch`, so the persisted artifact at
`.uf/artifacts/dispatch/<correlation>.json` carries findings,
advisories, plan, and runs but no proposals. The #929 design
accepted that loss and named a Phase 4 narrow read tool as the
remedy.

Five commands store lessons. `uf.review-pr`, `uf.triage-issue`, and
`uf.address-feedback` source them from a tool result;
`uf.review-council` from its Section 4 summary; `uf.unleash` Steps 4
and 6 finalize but never store dispatch proposals at all.

`uf.triage-issue` Section 2.3 and `uf.address-feedback` Section 2.3.7
retrieve a no-provenance `detail: "full"` payload because their
consolidation reads `runs`. The finalized artifact already persists
the same `runs` array (`payload.runs`, each with its
`workflow_verdict`).

## Goals / Non-Goals

### Goals

- Store every lesson with its full text.
- Store one lesson per idea when several personas submit the same
  one.
- Remove the unbounded full retrieval from every command.
- Keep artifacts written before this change valid.
- Keep the scaffold mirror byte-identical and the plugin suite at or
  above its 94 percent statements floor.

### Non-Goals

- No narrow read tool and no SQLite index (the #929 Phase 4
  follow-up stays open for the RD-FR-052 spill path).
- No change to the summary or compact caps (RD-FR-049, RD-FR-051).
- No change to verdict computation, finding deduplication, or run
  counting.
- No envelope or `schema_version` change.

## Decisions

### D1. Proposals live in the payload, not in a sidecar

`ReviewDispatchPayloadSchema` gains an optional `proposals` array of
`{agent, information, tag, category?}` entries: `agent` the
existing `AgentNameSchema`, `information` 1-16,384 characters, `tag`
1-128, `category` one of the five lesson categories, at most 512
entries, strict. `consolidate_dispatch` puts the deduplicated
proposals (D2) into the payload it finalizes and omits the key when
there are none. `artifactPayloadOrder` writes `proposals` after
`advisories` and before `plan` and `runs`, so a reader that pages
through the artifact reaches the lessons before the large arrays.

`schemas/review-dispatch/v1.0.0.schema.json` gains the same optional
property and a `lessonProposal` definition. The change is additive:
every artifact written before it still validates, so the
`schema_version` stays `1.0.0` and `parseReviewDispatchArtifact` in
`.opencode/lib/review-dispatch-artifact.ts` is untouched.

This amends RD-FR-050's rule that `proposals` never reach the
finalizer payload. `warnings` stay result-level only.

Alternatives rejected:

- **A sidecar file** (`<correlation>.proposals.json`). The
  `uf-workflow` resume scan reads every file in the flat dispatch
  store as an envelope, and consumers would need a second path. One
  artifact keeps one read.
- **Keep proposals under `runs[]`.** `DispatchRunSchema` is strict
  and per-run placement cannot express a cross-persona merge.
- **Cap `submit_lesson_proposal.information` at 512** (issue
  option 2). Personas write longer lessons because the lessons need
  the room; capping input moves the loss to the author.

### D2. Deterministic cross-persona deduplication

`deduplicateProposals(proposals)` runs once on the collected
proposals, before persistence, summary, full retrieval, and the
compact result, so every output agrees.

1. Tokenize `information`: Unicode NFKC, lowercase, split on runs of
   characters that are not letters or digits, and drop a fixed list
   of English stop words. Use the token set.
2. Two proposals are near-identical when the smaller token set has
   at least 4 tokens and the overlap coefficient
   `|A ∩ B| / min(|A|, |B|)` is at least 0.5, or when either set has
   fewer than 4 tokens and the two token sequences are equal.
3. Walk proposals in collection order. Compare each with the kept
   entries in order; on the first near-identical match keep the
   entry with the longer `information` (the earlier one on a tie)
   in the earlier slot, and drop the other.
4. Each drop appends a `duplicate_proposal_merged` warning
   `{type, agent, kept_agent}` (the dropped and kept entries'
   agents). The summary projection composes `detail` as
   `agent=<agent> kept=<kept_agent>`.

Tags and categories do not take part: the #941 duplicate pair used
the tags `tool-input-validation` and `opencode-plugins`. The
threshold comes from that session's 7 stored proposals: the
duplicate pair scores 0.59 and every other pair scores at most 0.18.

Alternatives rejected:

- **Exact normalized-text equality.** Misses the observed pair,
  which personas worded differently.
- **Embedding similarity.** Non-deterministic across model versions
  and needs a service the plugin does not have.
- **Merge into one entry with an `agents` list.** Changes the
  proposal shape every consumer and RD-FR-052 validate.

### D3. Tool-result proposals become a preview

The summary and compact proposal arrays keep their RD-FR-049 and
RD-FR-051 caps and `truncated` markers. They serve display and
host-truncation safety; the artifact is the only source a command
stores from. The tool description and every consumer say so.

### D4. Consumers read lessons from the artifact

After a successful finalization, each consumer reads the artifact at
`artifact_path` once (the read it already makes for findings and
advisories) and calls `dewey_store_learning` once per
`payload.proposals` entry. An absent `payload.proposals` means no
lesson was submitted. A failed finalization stores nothing; the
proposals stay in the dispatch store and are reported as stranded,
which is the rule `uf.review-pr` already had.

- `uf.review-council`: Section 4 keeps its summary retrieval and
  RD-FR-052 recovery for Section 3 presentation and warnings, but
  stores no lesson from it. Lessons are stored after Section 5
  finalization, from that iteration's artifact.
- `uf.review-pr`: Section 5 stores from the artifact instead of the
  compact result.
- `uf.unleash`: Steps 4 and 6 add lesson storage from the artifact
  they already read.

### D5 (revisits #929 D5). Triage and feedback finalize first

#929 kept `uf.triage-issue` and `uf.address-feedback` on
`detail: "full"` because their consolidation reads `runs`, and
relied on ordering to protect `runs`. Session
`ses_e622de6f4ffejp8CeJHa4cfTsV` shows full payloads of 42–51 KB for
ordinary PRs, and 4 of 7 were host-truncated; ordering cannot bound
the `runs` array itself.

Finalization does not depend on anything those commands compute
afterwards: `consolidate_dispatch` derives the payload from the
persisted dispatch session, and action outcomes are never part of
it. So both commands now run the provenance-bearing finalization
immediately after `dispatch_review_runs` and read `runs`, findings,
advisories, and proposals once from `artifact_path`:

- `uf.triage-issue`: Section 2.3 performs the Section 4.6
  finalization call (with its bounded mechanical retry) right after
  dispatch; Sections 2.4 and 3.1 and the `read_assessment_fields`
  run ids read the artifact. Section 4.6 documents the call and
  reports its result; it does not call again.
- `uf.address-feedback`: each Tier 2 item finalizes (Section 2.3.8)
  right after its dispatch; Sections 2.3.5 and 2.3.7 read that
  item's artifact.

The no-provenance retrieval and its truncation-check prose are
removed from both. A failed finalization leaves no artifact, so the
item or invocation fails closed (`INCONCLUSIVE`) before any GitHub
mutation; previously a late finalization failure could follow
mutations already made from retrieval data. The all-skip and
`not_found` semantics are unchanged.

Issue option 2 (spill recovery in both commands) was rejected: it
extends the DE-001 prose deviation instead of removing the
retrieval. Option 3 (a bounded `runs` projection in summary mode)
was rejected: the artifact already carries the complete `runs`.

With this change no in-repo command calls `consolidate_dispatch`
with `detail: "full"`. The mode stays the default for external
callers and for compatibility.

### D6. Spec deltas for `diff_not_found` (#943)

The `build-review-prompts` and `review-dispatch-diff-manifest`
deltas are unarchived, so their scenarios are edited in place:
FR-011 gains the in-scope missing-path rule with two scenarios,
FR-009's not-found scenario becomes a not-readable scenario (a
missing path never reaches the read), and RD-FR-034 gains a
`diff_not_found` scenario beside the existing
`path_validation_failed` one.

## Risks / Trade-offs

- **False-positive merges.** A 0.5 overlap could merge two distinct
  lessons on one topic. Every merge is reported as a warning naming
  both agents, the longer text survives, and the threshold sits
  well above the 0.18 measured between distinct lessons.
- **Artifact size.** 512 proposals of 16,384 characters would add
  about 8 MiB. Realistic dispatches carry fewer than 10 proposals
  of under 2 KB. Consumers already page through the artifact.
- **Older plugin, newer command.** A command reading an artifact
  written before this change finds no `payload.proposals` and
  stores nothing. The scaffold ships plugin and commands together.
- **Earlier finalization in triage and feedback.** A failed
  finalization now blocks the run before mutations. That is the
  safer order; the bounded mechanical retry is unchanged.
- **Coverage floor.** New branches are small and covered. If the
  94 percent floor fails, the Gatekeeping rule applies: report and
  stop.

## Archive order

This change amends requirements introduced by the unarchived
`bound-consolidate-dispatch-output` change (RD-FR-049 through
RD-FR-053), which in turn must archive after or with
`dispatch-finalize-determinism`. Archive this change in the same
batch as, or after, `bound-consolidate-dispatch-output`.
