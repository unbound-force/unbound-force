## Why

PR #929 (`bound-consolidate-dispatch-output`, issue #922) bounded
every `consolidate_dispatch` output so proposals and warnings
survive host truncation. Two of its decisions do not hold up in
live sessions.

**Lessons are stored cut to 512 characters (#941).** The compact
finalization result and the summary retrieval cut each proposal's
`information` to 512 characters (`SUMMARY_LIMITS.informationChars`).
The persisted artifact does not keep proposals, so the cut copy is
the only copy, and every consumer stores exactly that copy with
`dewey_store_learning`. `submit_lesson_proposal` accepts up to
16,384 characters, so personas routinely write longer lessons.

- Session `ses_e5dc3b9ccffevg6G3PMguRoAe3` (`/uf.review-pr 937`):
  all 7 proposals came back `truncated: true` and all 7 were
  stored at exactly 512 characters, ending mid-word ("…The decla",
  "…cannot c"). The final report said "7 lessons stored to Dewey".
- In the same run, adversary and architect each submitted the
  lesson "the host does not enforce the tool's declared args
  shape" under different tags, and both were stored.

**Full retrieval is host-truncated and fails closed (#945).**
`/uf.address-feedback` Section 2.3.7 and `/uf.triage-issue` Section
2.3 retrieve each dispatch with an unbounded no-provenance
`detail: "full"` call. Since #929 their truncation check fails
closed when the host truncates. Design D5 of #929 kept them on full
mode because they reason over `runs`; ordering alone cannot bound
that array.

- Session `ses_e622de6f4ffejp8CeJHa4cfTsV` (`/uf.address-feedback
  314` in unbound-force/gaze): 7 full retrievals of 42–51 KB
  each; the host truncated 4 of the 7. Under current prose those
  4 feedback items fail closed.

**Spec drift (#943).** PR #937 (#932) added `diff_not_found` to the
`DiffPathError` union of `read_diff_chunk` and
`build_review_prompts`; the capability deltas still list only
`path_validation_failed`.

Fixes #941, #945, #943.

## What Changes

- Persist the lesson proposals in the dispatch artifact:
  `consolidate_dispatch` adds an optional `payload.proposals`
  array (full text, `{agent, information <= 16384, tag <= 128,
  category?}`) to the payload it finalizes. The artifact orders it
  after `advisories` and before `plan` and `runs`. The strict Zod
  payload schema and `schemas/review-dispatch/v1.0.0.schema.json`
  gain the optional property; existing artifacts stay valid.
- Deduplicate near-identical proposals across personas before
  they are persisted or returned: a deterministic Jaccard rule with
  a negation guard keeps one entry per lesson and reports each
  merge as a new `duplicate_proposal_merged` warning. At most 512
  proposals persist; the rest are reported as `proposals_overflow`.
- Add `read_dispatch_artifact`, a bounded paged reader of the
  artifact by correlation ID, because the file-read tool cuts the
  artifact's long lines at 2,000 characters. Every consumer reads
  the artifact through it.
- The compact finalization result and the summary retrieval keep
  their bounded proposal arrays as a preview. No command stores
  lessons from them any more.
- Move every consumer onto the artifact:
  - `uf.review-pr.md`, `uf.unleash.md` Steps 4 and 6, and
    `uf.review-council.md` store lessons from the artifact's
    `payload.proposals` after finalization.
  - `uf.triage-issue.md` and `uf.address-feedback.md` finalize
    immediately after dispatch and read `runs`, findings,
    advisories, and proposals once from `artifact_path`. Their
    no-provenance `detail: "full"` retrieval is removed (revisits
    #929 design D5).
- Record `diff_not_found` in the `build-review-prompts` FR-009 and
  FR-011 scenarios and the `review-dispatch-diff-manifest`
  `read_diff_chunk` scenarios.
- No quality or governance gate changes. The plugin suite keeps
  its 94 percent statements floor.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: durable artifact proposals (RD-FR-054),
  cross-persona proposal deduplication (RD-FR-055), an
  artifact-sourced consumer contract (RD-FR-056), and a bounded
  artifact reader (RD-FR-057). Each amends
  requirements of the unarchived `bound-consolidate-dispatch-output`
  change (RD-FR-049 through RD-FR-053).
- `build-review-prompts` (FR-009, FR-011) and the
  `review-dispatch-diff-manifest` `read_diff_chunk` requirement
  (RD-FR-034): `diff_not_found` scenarios, edited in place in
  those unarchived deltas.

### Removed Capabilities
- None.

## Impact

- **Live plugin**: `.opencode/plugins/review-dispatch/index.ts` —
  `ReviewDispatchPayloadSchema` (optional `proposals`),
  `artifactPayloadOrder`, proposal deduplication and its warning,
  the warning projection, and the tool description.
- **Schema**: `schemas/review-dispatch/v1.0.0.schema.json` and its
  `.uf/schemas/` and scaffold copies gain the optional `proposals`
  property and a `lessonProposal` definition.
- **Commands**: `uf.review-council.md` Sections 2, 4, and 5;
  `uf.review-pr.md` Section 5; `uf.unleash.md` Steps 4 and 6;
  `uf.triage-issue.md` Sections 2.3, 2.4, 3.1, and 4.6;
  `uf.address-feedback.md` Sections 2.3.4, 2.3.5, 2.3.7, and
  2.3.8.
- **Scaffold mirror**: every edited `.opencode/` file is copied
  byte-identical under `internal/scaffold/assets/opencode/`.
- **Tests**: vitest coverage for persistence, ordering, and
  deduplication; Go content gates in `internal/scaffold/` and a
  schema test in `internal/schemas/`.
- **Behavior change**: proposals are durable and complete in the
  artifact. A finalization failure now stores no lessons (they stay
  in the dispatch store, as `uf.review-pr` already documented).
  Triage and feedback consolidation no longer read a retrieval
  payload, so a failed finalization fails those items closed
  before any mutation.
- **CI replication**: `make plugin-test` and
  `go test -race -count=1 ./...`.
- **Documentation impact**: CHANGELOG.md entries. The artifact
  shape and consumer behavior are user-facing to calling agents,
  so the documentation-gate issue is filed before merge.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The artifact becomes the one self-describing record of a dispatch,
lessons included. Consumers read a file the producer already wrote
instead of re-running a retrieval that the host may truncate.

### II. Composability First

**Assessment**: PASS

The change stays inside the review-dispatch plugin, its schema, and
its command consumers. The new payload property is optional, so
artifacts written before this change remain valid.

### III. Observable Quality

**Assessment**: PASS

Lessons are persisted with provenance (`agent`) beside the findings
they came from. Every merged duplicate is reported as a structured
warning, so deduplication is never silent.

### IV. Testability

**Assessment**: PASS

Deduplication is a pure function over the collected proposals.
Persistence and ordering are asserted through the existing injected
finalization dependencies with no external service.

### V. Security by Default

**Assessment**: PASS

No new input surface: proposals were already Zod-validated at
submit and harvest time and are validated again by the strict
payload schema. Consumers read only the artifact path the tool
returned, inside `.uf/artifacts/dispatch/`.

### VI. Deterministic Execution

**Assessment**: PASS

Deduplication runs in the tool, not in command prose. Removing the
full retrieval from two commands removes a model-executed
truncation-handling branch.
