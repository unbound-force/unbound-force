## Why

`consolidate_dispatch` returns its entire result as one unbounded,
pretty-printed JSON string. The two workflow-critical arrays —
`proposals` and `warnings` — are appended LAST, after the payload
body. opencode truncates tool output host-side at roughly 50 KiB,
so those arrays are the first bytes cut. `uf.review-council.md`
Section 4 then requires a single retrieval call and says "do not
re-fetch it", with no truncation handling. That pushes a
deterministic step (parse the JSON, iterate `proposals`, validate
each entry) onto LLM improvisation — a Determinism-gate violation
(DE-001/DE-005) and a silent data-loss path for lesson capture.

Observed failure (review session
`ses_e62531039ffeTRy54bzgRkZodZ`, dispatch correlation
`1ee21684-38cb-42d1-9550-08da5c9ee0c8`):

- Retrieval call: full payload 63,915 bytes; the host kept 51,070
  characters and truncated 12,871 bytes — the tail of `findings`
  plus the entire `proposals` array (7 proposals).
- Finalization call: 210,564 bytes returned, 159,569 truncated. It
  echoes `review_dispatch` and `calculated_assessment` even though
  the authoritative artifact is already persisted at
  `artifact_path` (`.uf/artifacts/dispatch/<correlation>.json`).

The same unbounded shape feeds every consumer of
`consolidate_dispatch`, so the defect is structural, not a single
command's oversight.

Fixes #922.

## What Changes

- Add a compact retrieval mode to `consolidate_dispatch`:
  `detail: "summary"` returns only `correlation_id`, `proposals`,
  `warnings`, `verdict`, `verdict_reason`, `run_counts`, `coverage`,
  `findings`, `advisories`, and four truncation flags —
  observability first, so a truncated tail cannot lose it. The
  summary caps findings and advisories at 15 entries each with a
  slim projection (category ≤128, description ≤512, file ≤256;
  `root_cause` and `run_ids` stay in the artifact), proposals at 10
  (agent ≤128, information ≤512, tag ≤128), warnings at 5 (kind ≤64,
  detail ≤256, composed per kind from the structured warning), and
  truncates `verdict_reason` to 512 characters. It sets the matching
  `findings_truncated`, `advisories_truncated`,
  `proposals_truncated`, or `warnings_truncated` flag and marks each
  shortened entry with `truncated: true`. After assembly it then
  serializes compactly, measures the UTF-8 byte length, and drops
  trailing entries in the fixed order findings, advisories,
  proposals, warnings until the result is strictly under 40,960
  bytes. The default stays `full` for backward compatibility.
- Reorder the full retrieval payload so `correlation_id`,
  `proposals`, and `warnings` precede `verdict` and the large
  arrays, keeping `runs` ahead of `plan`, `findings`, and
  `advisories`, as cheap insurance for any consumer that still
  receives a truncated payload.
- Compact the finalization response to `status`,
  `operation_verdict`, `correlation_id`, `artifact_path`, scalar
  `review_verdict`, `run_counts`, bounded `proposals` / `warnings`
  (10 / 5 entries with their flags), and a body-free
  `calculated_assessment: { authoritative }` flag, followed by
  `errors`. Stop echoing `review_dispatch` and the payload body.
- Make `uf.review-council.md` Section 4 deterministic: use the
  compact mode and specify an exact spill-file fallback (read the
  host-saved output path, extract `proposals` / `warnings`,
  validate) instead of "do not re-fetch" plus ad-hoc explore
  delegation.
- Update every command that consumes the compacted result
  (`uf.review-pr.md`, `uf.unleash.md` Steps 4 and 6,
  `uf.triage-issue.md`, `uf.address-feedback.md`) so no consumer
  reads a field the tool no longer returns; findings and advisories
  come from the persisted artifact at `artifact_path`.
- Keep `uf.triage-issue.md` Section 2.3 and
  `uf.address-feedback.md` Section 2.3.7 retrieval in `full` mode
  because their consolidation reads the `runs` array; the full-mode
  reordering already protects proposals and warnings.
- Update `.opencode/test/golden-output-shape.test.ts`,
  `.opencode/test/consolidate-dispatch.test.ts`, and
  `.opencode/test/finalize-error-boundaries.test.ts`. Add a
  size-budget assertion (a 16-finding / 7-proposal dispatch
  retrieval stays under 40 KiB (40,960 bytes), guaranteed by the
  post-assembly drop loop), a multibyte and escape-heavy fixture,
  and proposals-before-findings ordering tests.
- No quality or governance gate is weakened. The plugin suite keeps
  its 94 percent statements coverage floor.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `review-dispatch`: `consolidate_dispatch` gains a bounded
  summary retrieval mode (RD-FR-049), a transport-safe full-payload
  ordering (RD-FR-050), a compact finalization response that stops
  echoing the artifact body (RD-FR-051), and a deterministic
  retrieval-recovery contract for the review-council command
  (RD-FR-052), and a consumer command contract for every command
  that calls the tool (RD-FR-053).

### Removed Capabilities
- None.

## Impact

- **Live plugin**: `.opencode/plugins/review-dispatch/index.ts` —
  `ConsolidateDispatchInputSchema` (add `detail`), the payload
  assembly and both `JSON.stringify` return sites in
  `createConsolidateDispatchTool`, and the tool description.
- **Scaffold mirror**:
  `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts`
  — byte-identical copy enforced by the Go drift gate
  `TestEmbeddedAssets_MatchSource`.
- **Commands**: `.opencode/commands/uf.review-council.md` Section 4
  (retrieval) and Section 5 (finalization result consumption);
  `.opencode/commands/uf.review-pr.md` Section 5;
  `.opencode/commands/uf.unleash.md` Steps 4 and 6;
  `.opencode/commands/uf.triage-issue.md` Sections 2.3 and 4.6;
  `.opencode/commands/uf.address-feedback.md` Sections 2.3.7 and
  2.3.8. Each command mirrors byte-identically into
  `internal/scaffold/assets/opencode/commands/`.
- **Tests**: `.opencode/test/golden-output-shape.test.ts` (new
  compact shapes, ordering, size budget),
  `.opencode/test/consolidate-dispatch.test.ts` (summary mode,
  default compatibility, compact finalization), and
  `.opencode/test/finalize-error-boundaries.test.ts` (compact
  result carries no artifact body).
- **Behavior change**: callers that read the finalization result's
  `review_dispatch`, `calculated_assessment.payload`, or object
  `review_verdict` must read the persisted artifact at
  `artifact_path` and the scalar decision instead. This is the
  intended size reduction.
- **CI replication**: `make plugin-test` (vitest with the 94
  percent statements floor) and
  `go test -race -count=1 ./internal/scaffold/...` (drift gate).
- **Out of scope**: a write-once/read-many local store with narrow
  read tools, and a SQLite index. Both are possible follow-ups.
- **Documentation impact**: CHANGELOG.md entry; the change is
  user-facing to calling agents and operators (tool output shape),
  so the Cross-Repo Documentation rule applies and the in-repo and
  unbound-force/website documentation issues are filed before PR
  merge. Content coverage: no blog or tutorial issue is warranted —
  this is an output-contract fix for internal agent commands, not a
  user-visible feature a reader would follow.

## Constitution Alignment

Assessed against the Unbound Force org constitution (v1.4.0 — six
active principles).

### I. Autonomous Collaboration

**Assessment**: PASS

The retrieval result becomes bounded and self-describing: verdict,
counts, proposals, and warnings survive host transport, so the
parent command can act without improvising a second retrieval path
or delegating exploration to another agent. No synchronous coupling
is introduced.

### II. Composability First

**Assessment**: PASS

The change stays inside the `review-dispatch` plugin and its
command consumers. It adds no dependency, no new service, and no
cross-hero requirement. Each hero repo remains independently
installable, and the scaffold mirror keeps deployed copies identical.

### III. Observable Quality

**Assessment**: PASS

The compact finalization response keeps the authoritative artifact
path, the scalar review decision, run counts, the assessment
authority flag, and errors, so every decision remains
machine-parseable with provenance. Removing the echoed artifact
body improves signal density: the artifact at `artifact_path`
remains the single source of truth.

### IV. Testability

**Assessment**: PASS

Both shapes are produced by the existing tool and asserted through
the vitest suites with injected filesystem dependencies. The
size-budget and ordering tests are pure assertions over the tool's
string output; no external service or network is required.

### V. Security by Default

**Assessment**: PASS

No new credentials, external services, or unvalidated input
surface are added. The `detail` field is closed-validated by the
existing strict schema, and the spill-path recovery accepts a path
only from the strict host-notice shape inside the opencode
tool-output directory. The compact result redacts nothing that was
previously protected — it simply stops re-emitting content already
persisted under the artifact path.

### VI. Deterministic Execution

**Assessment**: PASS WITH AUTHORIZED DEVIATION (DE-001)

The primary mitigation is tool-rendered and size-bounded: retrieval
moves from an unbounded string to a capped summary, and finalization
stops echoing the artifact body. The recovery path itself is
model-executed prose because no narrow read tool exists at this
revision; the deviation is authorized for this change, bounded by
the strict notice shape, directory confinement, non-symlink check,
1 MiB size cap, and content validation in RD-FR-052. Trigger for
extraction: the fallback firing in more than one observed
`uf.review-council` run per month. Owner: the review-dispatch
maintainer. Phase 4 files the follow-up issue for the narrow read
tool and records it in the PR body.

A second accepted trade-off is content loss. The summary and the
compact finalization result cap proposals at 10 and warnings at 5
and shorten individual fields, marking each shortened entry with
`truncated: true`. Entry drops are signalled by the four top-level
flags; field shortening is signalled per entry. Neither a dropped
nor a shortened proposal or warning is recoverable from the
RD-FR-052 spill (the spill is the same capped output) or from the
artifact (which persists findings and advisories only). This loss
is accepted for this change and remedied by the Phase 4 narrow read
tool.

**Archive order.** The two cross-capability requirements this change
touches — `dispatch-finalization` FR-001 and the review-council
"Finalize Every Iteration Via Consolidation" requirement — live in
the unarchived `dispatch-finalize-determinism` change, so
`openspec/specs/` cannot receive MODIFIED deltas for them yet. That
change MUST archive before, or in the same batch as, this one.
RD-FR-051 is the reciprocal reference: it defines the compact
finalization result as authoritative for the outcome and scalar
decision while the full projection is read once from
`artifact_path`.
