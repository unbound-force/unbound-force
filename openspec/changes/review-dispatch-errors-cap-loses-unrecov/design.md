## Context

`buildCompactFinalization` (`.opencode/plugins/review-dispatch/index.ts`)
projects the finalizer's `errors` for the compact result with
`SUMMARY_LIMITS.errors` (5 entries) and
`SUMMARY_LIMITS.errorChars` (256 characters per string). The
finalizer's three failure paths produce errors of very different
shapes:

- `structuralFailure` — one string composed by
  `formatValidationError` (`.opencode/lib/reviewer-manifest.ts`),
  which joins every Zod issue of the rejected input with `"; "`.
  At roughly three issues the composition passes 256 characters;
  the projection cuts it mid-string.
- semantic rejection (`validatePayloadSemantics`) — one short
  string per violation; per-finding and per-run checks over
  payloads of up to 512 findings and 64 runs produce dozens.
- persistence failure — one string carrying the filesystem
  error.

On all three paths `artifact_path` is null: no artifact is
written, and `read_dispatch_artifact` (RD-FR-057) has no
`errors` section. Whatever the preview cuts or drops is
unrecoverable — the gap @uf-divisor flagged as MEDIUM (sre) on
PR #1004, filed as issue #1039. The identical code ships in the
scaffold mirror under `internal/scaffold/assets/`.

RD-FR-051 recorded the observability caps as an accepted risk
with the narrow read tool as remedy; RD-FR-058
(`full-lesson-proposals`) removed mid-string cuts for proposals
because the persisted artifact made the preview safely lossy.
Errors never received either remedy.

## Goals / Non-Goals

### Goals

- No compact finalization result can hand a consumer a
  shortened error string.
- Every dropped preview entry is recoverable from durable state
  in the project worktree (`.uf/artifacts/dispatch/`), readable
  through `read_dispatch_artifact`.
- Every tool result stays under the 40,960-byte budget
  (RD-FR-049) without weakening it.
- The fix lands byte-identical in the scaffold mirror.

### Non-Goals

- No change to `formatValidationError`'s composition: the
  complete string is the diagnostic value; completeness is the
  point, and every consumer of the function benefits.
- No change to the findings, advisories, warnings, or
  `verdict_reason` caps (RD-FR-049 / RD-FR-058 govern them).
- No retry loop or recovery workflow for failed finalizations
  in commands: the record makes failures diagnosable; fixing
  them stays with the operator (epic #1060 handles retry
  behavior separately where scoped).
- No `dispatch_status` changes: the session store is
  tmpdir-scoped observability, not the durable recovery
  surface this change adds.

## Decisions

### D1. Errors are whole or absent in the preview

`SUMMARY_LIMITS.errorChars` is removed; the projection copies
each error string verbatim. The 5-entry count cap stays, and
`errors_truncated` is true only when whole entries were dropped
(mid-string shortening no longer exists, so the flag's
`some(truncated)` disjunct goes away). This is the RD-FR-058
present-whole-or-absent pattern applied to errors. Unlike
proposals, error strings have no schema bound at their source,
which is why D3's record and D2's budget are required to keep
the result bounded.

Rejected: raising `errorChars` (e.g. to 4,096). Any fixed
per-string cap re-creates the original defect at a different
scale — the structural composition is unbounded in issue count,
so a "sized" cap is still a mid-string cut for some input.

### D2. Errors join the byte-budget loop, dropping after
proposals

The compact builder's post-assembly budget loop currently drops
only proposals and exits once they are exhausted, because every
other array is per-entry capped and therefore bounded by
construction. Uncapped errors break that argument, so the loop
now drops trailing errors whole after proposals are exhausted,
setting `errors_truncated`. Errors drop last because they are
the diagnostic payload of exactly the paths where nothing else
is persisted; proposals are recoverable from the success
artifact when the result is a success, and on failure paths the
two rarely compete (errors only exist on failure paths). The
40,960-byte budget itself is untouched (Gatekeeping rule: the
budget is an RD-FR-049 governance constant).

### D3. A durable failure record per failed finalization

Each failed `finalizeReviewDispatch` best-effort writes a
failure record to
`.uf/artifacts/dispatch/<correlation-id>.failure.json`:

- **Envelope**: the Hero Interface Contract shape already used
  for success artifacts — `hero: "the-divisor"`, producer
  version, timestamp, `artifact_type: "review-dispatch-failure"`,
  schema version, `context { branch, commit, correlation_id,
  workflow_id }`, payload. The distinct `artifact_type` keeps
  resume detection (`parseReviewDispatchArtifact` in
  `.opencode/lib/review-dispatch-artifact.ts`) ignoring the
  record — that consumer is fail-open on foreign types by
  design.
- **Identity**: the record's correlation id is the dispatch's
  when a schema-usable one exists; otherwise a fresh UUID from
  the injected `uuid` dependency (structural rejections of
  garbage inputs have no usable id). The payload always carries
  `dispatch_correlation_id` (null when unusable) plus
  `command`/`mode` when known, `status: "failed"`,
  `operation_verdict`, and the complete `errors` verbatim.
- **Publish discipline**: identical to `persistEnvelope` —
  0o750 directory, exclusive-create temp file, rename-by-link,
  cleanup on failure, and `EEXIST` collision retry that mints a
  fresh record correlation id (a re-failed dispatch appends a
  new record rather than overwriting the old one; both remain
  discoverable from their respective compact results).
- **Bounds**: the record's strict schema bounds every string
  with sanity limits sized above the natural composition
  ceilings — per entry 131,072 characters (the structural
  composition ceiling is roughly 100 KiB: one issue per invalid
  field path over a schema of bounded arrays, each Zod message
  short) and 65,536 entries (the semantic ceiling is roughly
  34,000 short strings from the contributors/consolidated
  cross-checks at the payload's 64-run/512-finding maxima).
  The bounds fail closed and never truncate: a composition that
  somehow exceeded them would leave no record (D4 degradation)
  rather than a silently damaged one.

The result surfaces the record through one new compact-result
key, `failure_record` — `null` or `{ correlation_id, path }` —
omitted when null and inserted immediately after
`artifact_path` (early keys survive host tail-truncation;
RD-FR-051's order is amended in the delta). `failure_record`
was chosen over reusing `artifact_path` to keep that key's
meaning ("the persisted review-dispatch artifact") unchanged
for existing consumers.

`.uf/artifacts/` is gitignored runtime data, so records follow
the same lifecycle as success artifacts — never committed,
bounded by failure count, small (kilobytes in practice).

### D4. The record write never masks the failure it records

The write is wrapped so a record-write failure cannot replace
or hide the finalization failure: the result returns
`failure_record: null` and appends one bounded
`persist failure record: <message>` entry to the result's
`errors` (which the D2 budget treats like any other entry). On
the persistence-failure path the same environmental cause may
fail both writes (disk full); the degradation is exactly the
pre-change behavior plus whole-entry previews, so the change
strictly improves every path it touches.

### D5. `read_dispatch_artifact` section `errors`

The section enum gains `errors`. The reader derives
`.uf/artifacts/dispatch/<correlation_id>.failure.json` from the
caller-supplied correlation id (no caller-supplied paths),
validates the regular-file/64 MiB/parse/type/payload chain with
the existing RD-FR-057 error codes, and returns
`{ status: "ok", correlation_id, section, total, errors }` with
the complete array. A dispatch that finalized successfully has
no record: the section returns `not_found` — success artifacts
carry no errors by schema, so the answer is authoritative, not
a probe ordering.

### D6. Schema registry and scaffold mirrors

The record payload is declared in a hand-authored
`schemas/review-dispatch-failure/v1.0.0.schema.json`, registered
in `internal/schemas` fixture tests
(`handAuthoredSchemas`), and mirrored under
`.uf/schemas/review-dispatch-failure/` and
`internal/scaffold/assets/uf/schemas/review-dispatch-failure/`
with the scaffold asset list in `scaffold_test.go` gaining the
path. Every edited `.opencode/` file is copied byte-identical
under `internal/scaffold/assets/`, keeping the existing drift
detection (`review-dispatch index.ts` and `reviewer-manifest.ts`
are already in the pinned asset list) green.

### D7. Prose corrections only

The `consolidate_dispatch` description sentence "errors (capped
at 5 strings of 256 characters, flagged by errors_truncated)"
now describes complete-or-dropped entries and the
`failure_record` key. The `read_dispatch_artifact` description
gains `errors` to its section list. No command workflow prose
changes and no mechanical steps are added anywhere (DE-001).

## Risks / Trade-offs

- **Record accumulation.** Repeated failed finalizations append
  records (never overwrite). Records are kilobytes in the
  gitignored runtime directory with the same growth model as
  success artifacts; accepted.
- **Dual-write failure.** On the persistence-failure path both
  writes can fail together; the degraded result still carries
  whole preview entries (D4). Accepted — strictly better than
  the status quo.
- **Pathological compositions.** An error composition beyond
  the D3 sanity bounds leaves no record (fail closed). The
  ceiling analysis makes this unreachable from real payloads;
  the bound exists so absurd input cannot produce an unbounded
  file.
- **Coverage floor.** The plugin suite carries a 94 percent
  statements floor and the review-dispatch gate requires 85
  percent branch coverage; the new error paths (three failure
  paths, record-write failure, reader section) are exactly the
  kind that dropped coverage in past changes. Tests are planned
  per path from the start; if a floor fails, the Gatekeeping
  rule applies.

## Constitution alignment

Assessed in the proposal (all six principles PASS). Decisions
D3 and D5 are the Observable Quality and Autonomous
Collaboration core: the failure record is a self-describing,
machine-parseable artifact with provenance, discoverable by any
consumer without the producing session. D2 and D6 preserve
Composability First — no new runtime coupling, the resume
consumer's fail-open contract is untouched, and mirrors keep
scaffolded deployments byte-equal.

## Archive order

This change amends clauses of the unarchived
`bound-consolidate-dispatch-output` (RD-FR-051) and
`review-dispatch-consumer-bounds` (RD-FR-057) changes, and must
respect `full-lesson-proposals` (RD-FR-058) as the
whole-or-absent precedent it extends. Archive it in the same
batch as, or after, both amended changes.
