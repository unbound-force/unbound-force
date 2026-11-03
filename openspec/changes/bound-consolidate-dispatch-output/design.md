## Context

`createConsolidateDispatchTool`
(`.opencode/plugins/review-dispatch/index.ts`, L4167) builds one
`payload` object, then returns `JSON.stringify({ ...payload,
proposals?, warnings? }, null, 2)` at two sites:

- the provenance branch (L4509-4523), which spreads
  `FinalizeReviewDispatchResult` and appends `proposals` /
  `warnings`;
- the no-provenance branch (L4526-4535), which returns the full
  payload with the observability fields appended.

Both are unbounded. The finalization result also echoes
`review_dispatch` (the whole persisted envelope) and
`calculated_assessment` (which embeds the payload again), so a
210,564-byte response is mostly redundant body. The retrieval call
loses `proposals` first because they are appended last.

The parent command's Section 4 currently says: call
`consolidate_dispatch` once without provenance, use the top-level
`proposals` array, and "do not re-fetch it". Nothing tells the
agent what to do when the host truncates. Section 5 reads
`artifact_path`, object `review_verdict`, and
`calculated_assessment.authoritative`. `uf.unleash.md` Steps 4 and 6
and the sibling finalization sections (`uf.review-pr.md`,
`uf.triage-issue.md`, `uf.address-feedback.md`) read the same fields.

## Goals / Non-Goals

### Goals

- Keep the retrieval payload bounded so `proposals` and `warnings`
  survive host transport.
- Preserve the default full shape for backward compatibility.
- Make the finalization response proportional to the decision, not
  to the artifact body.
- Replace the "do not re-fetch" instruction with an exact,
  bounded recovery procedure.
- Keep every consumer command consistent with the new shapes.
- Keep the plugin suite at or above the 94 percent statements
  floor and the scaffold mirror byte-identical.

### Non-Goals

- No write-once/read-many local store and no SQLite index. Both
  are follow-ups.
- No change to verdict computation, finding deduplication, run
  counting, or artifact persistence.
- No change to `finalize_review_dispatch`'s own tool result; the
  compaction applies to `consolidate_dispatch`'s wrapper.
- No change to the strict `ReviewDispatchPayloadSchema`.
- No new dependency.

## Decisions

### D1. `detail: "summary"` is a retrieval-only input

Add `detail: z.enum(["full", "summary"]).default("full")` to
`ConsolidateDispatchInputSchema` and keep `.strict()`. A
`superRefine` rejects `detail: "summary"` combined with
`provenance`: finalization always returns the compact result of D3,
so the combination is ambiguous and fails closed instead of being
silently ignored.

Rationale: default `full` preserves every existing caller,
including tests and the sibling commands before they are updated.
An explicit summary mode keeps the bounded path opt-in at the tool
boundary while commands adopt it deliberately.

The summary builder emits exactly the RD-FR-049 field set by
initializing a fresh object literal in this order: `correlation_id`,
`proposals?`, `warnings?`, `verdict`, `verdict_reason`, `run_counts`,
`coverage`, `findings`, `advisories`, `findings_truncated?`,
`advisories_truncated?`, `proposals_truncated?`, `warnings_truncated?`.
It never spreads `payload`, so `plan`, `runs`, and the context fields
cannot leak in by accident, and it never emits `artifact_path`
because a retrieval persists nothing. The builder projects and caps
entries so the result is bounded independently of the underlying
payload: `findings` at 15 entries, each projected to `{severity,
category <= 128, description <= 512, file <= 256 or null, line}`
with `root_cause` and `run_ids` dropped; `advisories` at 15 entries,
each projected to `{severity, description <= 512}` with `run_ids`
dropped; `proposals` at 10 entries with `agent` truncated to 128,
`information` truncated to 512 and `tag` to 128; `warnings` at 5
entries, each projected to `{kind <= 64, detail <= 256}` with
`detail` composed per kind; and `verdict_reason` is truncated to
512 characters. An included entry carries `truncated: true` when
any of its emitted fields was shortened. The four top-level flags
fire when an entry is dropped. After assembly the builder
serializes the summary compactly (no indentation), measures its
UTF-8 byte length, and while the length is at or above 40,960
drops trailing entries in the fixed order findings, advisories,
proposals, warnings (setting the matching flag) until the result
is strictly under 40,960; a scalar-only summary is already below
budget, so the loop terminates. The complete `findings` and
`advisories` stay in the persisted artifact. A dropped or
shortened proposal or warning is not recoverable from the
RD-FR-052 spill, because that spill is the same capped summary; it
is an accepted content-loss trade-off until the Phase 4 narrow
read tool lands.

Alternatives rejected:

- **Bound the full payload with a byte cap and truncate strings.**
  Truncation would still cut arrays and silently change findings;
  the caller needs complete findings and proposals.
- **Make summary the default.** Breaks compatibility with existing
  consumers and tests before they are migrated.
- **Use provenance presence to imply summary.** Conflates two
  orthogonal axes (finalize vs retrieve, full vs summary).

### D2. Explicit transport ordering for the full payload

The finalizer payload and the retrieval output are built
separately. `payload` remains the exact object passed to
`finalizeReviewDispatch`. The full retrieval result is rebuilt from
`payload` in this insertion order:

```
command, mode, full, input_context, change_profile, plan_version,
correlation_id, proposals?, warnings?, verdict, verdict_reason,
run_counts, coverage, runs, plan, findings, advisories, workflow_result
```

`JSON.stringify` preserves insertion order, so observability fields
serialize before the dispatch body and `runs` stays ahead of `plan`
and the finding/advisory arrays. The values are the same objects;
only ordering changes. `proposals` and `warnings` are omitted when
empty. The strict finalizer payload never receives `proposals` or
`warnings`.

Alternatives rejected:

- **Post-process the JSON string to move keys.** String surgery is
  brittle and risks corrupting escaped content.
- **Add a replacer function.** Key ordering via a replacer is
  non-obvious and harder to test than an explicit object literal.

### D3. Compact finalization response with one scalar assessment flag

`compactFinalizationResult(result)` maps
`FinalizeReviewDispatchResult` to:

```
status, operation_verdict, correlation_id, artifact_path,
review_verdict, run_counts, proposals?, warnings?,
proposals_truncated?, warnings_truncated?,
calculated_assessment, errors
```

`proposals` and `warnings` are bounded before they are appended: at
most 10 proposals with `information` truncated to 512 characters and
`tag` to 128, and at most 5 warnings with `kind` <= 64 and `detail`
<= 256, each shortened entry marked `truncated: true`; the matching
top-level flag is set when an entry is dropped. Observability keys
therefore precede `calculated_assessment` and `errors`, so a host cut
cannot remove proposals or warnings before the outcome fields. A
dropped or shortened proposal or warning is not recoverable from the
artifact (which persists findings and advisories only); this is an
accepted risk recorded beside the DE-001 deviation, remedied by the
Phase 4 narrow read tool.

- `review_verdict` is the scalar `council_decision` from
  `result.review_verdict`, falling back to
  `result.calculated_assessment.review_verdict.council_decision`
  for a persistence failure, else null.
- `run_counts` is `result.calculated_assessment.payload.run_counts`
  when a payload exists, else null.
- `calculated_assessment` is null or `{ authoritative: boolean }`.
  The payload body and nested verdict are dropped.

Keeping `{ authoritative }` is a deliberate refinement of the
issue's field list: the command's bounded mechanical-retry
classification depends on distinguishing a validation rejection
(null) from a persistence failure (`authoritative: false`), and the
determinism pack forbids inferring that from error prose (DE-006).
The payload itself is the bloat, not the boolean.

`review_dispatch` is dropped entirely; the authoritative artifact
stays at `artifact_path`.

Alternatives rejected:

- **Keep `calculated_assessment` intact for failure context.**
  That is the 200 KB echo the issue targets.
- **Drop `calculated_assessment` entirely.** Would force the
  command to pattern-match `errors` text, violating DE-006.
- **Rename the flag to a new top-level field.** A new name would
  require parallel command edits for no behavioral gain; keeping
  the established `.authoritative` path minimizes consumer churn.

### D4. Section 4 uses summary and an exact spill-file fallback

`uf.review-council.md` Section 4 changes to:

1. Call `consolidate_dispatch` without provenance with
   `correlation_id` and `detail: "summary"`.
2. If the output contains no line matching the truncation-notice
   shape, treat it as complete and iterate `proposals`.
3. If it contains a matching notice line, accept the saved path only
   when the notice names exactly one path (double-quoted and possibly
   containing spaces, or unquoted) and the path is absolute, has no
   `..` segment, has no symbolic link at any component, resolves
   strictly beneath the realpath of the opencode tool-output
   directory on a `/` boundary (never equal to the directory itself),
   and is a regular file of at most 1 MiB; otherwise treat the
   retrieval as unrecoverable. The notice's second line (an
   explore-delegation recommendation) is ignored. For an accepted
   path, read it exactly once, discard any truncated leading fragment,
   parse the same JSON, extract `proposals` and `warnings`, and
   validate each entry against the RD-FR-052 shapes, then store.
4. If parsing or validation fails, or no acceptable path exists,
   fail closed without storing.

The section retries the identical summary call exactly once on any
tool error and keeps the prohibition on sourcing proposals from
`dispatch_review_runs`, child transcripts, or session-database
reads. It removes the unconditional "do not re-fetch" sentence.

Notice detection applies to every `consolidate_dispatch` response,
including `detail: "full"`. The two full-mode consumers that reason
over the `runs` array (`uf.triage-issue.md` and
`uf.address-feedback.md`) MUST treat a detected truncation notice as
a failure signal and fail closed instead of consolidating a partial
`runs` array; ordering alone cannot bound that array.

Determinism trade-off (DE-001), recorded as an authorized deviation
rather than clean compliance: the primary mitigation is tool-side
bounding (RD-FR-049 and RD-FR-051), so the fallback is exceptional,
but the fallback itself is model-executed prose and therefore not a
deterministic tool. The deviation is bounded three ways: the path is
accepted only from the strict host-notice shape and the confinement
rules of RD-FR-052, so the step cannot read an arbitrary file; the
parsed content is validated before use; and a tool-executable
alternative is not yet feasible because no narrow read tool exists at
this revision. Trigger for extraction: the fallback firing in more
than one observed `uf.review-council` run per month. Owner: the
review-dispatch maintainer. The follow-up narrow read tool is tracked
as its own issue filed in Phase 4; until it lands, the fallback stays
the documented interim mitigation.

### D5. All five consumers move to the compact shape

Because D3 removes `review_dispatch`, the object `review_verdict`,
and `calculated_assessment.payload`, every command that reads them
must move to `artifact_path`, the scalar `review_verdict`, and the
scalar `authoritative` flag. Findings and advisories come from the
persisted artifact at `artifact_path` (`payload.findings` and
`payload.advisories`), because the compact result no longer carries
them.

| Command | Retrieval | Finalization |
|---|---|---|
| `uf.review-council.md` | Section 4: `detail: "summary"` plus the RD-FR-052 recovery path | Section 5: compact result |
| `uf.review-pr.md` | none added: Section 5 already finalizes with provenance and forbids a separate no-provenance retrieval | Section 5: compact result; proposals and warnings from it, findings and advisories from `artifact_path` |
| `uf.unleash.md` | none: Steps 4 and 6 finalize with provenance | Steps 4 and 6: compact result; findings and advisories read once from `artifact_path` |
| `uf.triage-issue.md` | Section 2.3 stays `detail: "full"` because its rules read `runs` | Section 4.6: compact result |
| `uf.address-feedback.md` | Section 2.3.7 stays `detail: "full"` | Section 2.3.8: compact result |

The asymmetry is deliberate. `uf.review-council` is the only consumer
that retrieves proposals without finalizing, so it is the only one
that adopts summary retrieval. `uf.triage-issue` and
`uf.address-feedback` already retrieve full payloads to reason over
`runs`, and the full-mode reordering (D2) already moves proposals and
warnings ahead of `runs`. `uf.review-pr` and `uf.unleash` finalize
with provenance and must not add a second retrieval call. This is
consumer consistency required by the tool contract change, not a new
feature; RD-FR-053 makes it normative.

### D6. Size budget and ordering are asserted as parsed properties

The regression tests assert the byte size on the raw string and
the ordering on the parsed object:

- `Buffer.byteLength(output, "utf8") < 40 * 1024` for a summary
  result built from 16 unique findings and 7 proposals. Byte
  length is the only true transport check, so it stays on the
  raw string.
- Parsed key order via `Object.keys(JSON.parse(output))`: for the
  full result, `proposals` and `warnings` precede `runs`, and
  `correlation_id` precedes `plan`; for the summary result,
  `correlation_id` precedes `proposals`, `proposals` precedes
  `warnings`, and `warnings` precedes `findings`.
- A worst-case summary built from 512 findings, 512 advisories,
  512 proposals, and 512 warnings asserts the 15/15/10/5 caps,
  the 512-character description, information, and `verdict_reason`
  truncations, the 256-character file and warning-detail
  truncations, the 128-character category, tag, and agent
  truncations, all four truncation flags, and a total byte length
  strictly under 40,960. The guarantee is bounded by construction
  rather than derived arithmetic: the builder serializes the
  summary compactly, measures its UTF-8 byte length, and drops
  trailing entries in the fixed order findings, advisories,
  proposals, warnings (setting the matching flag) until the result
  is strictly under 40,960; a scalar-only summary is far below
  budget because `verdict_reason` is capped, so the loop
  terminates. The fixture additionally covers multibyte and
  escape-heavy text (CJK, astral emoji, quotes, backslashes, and
  control escapes), which a character-count derivation would
  understate.

Raw-string `indexOf` is not used for ordering: a key name can also
appear inside a value, so only `Object.keys` on the parsed object
proves insertion order.

### D7. Test placement

- `.opencode/test/consolidate-dispatch.test.ts`: summary field set,
  default-full compatibility, `summary` + provenance rejection,
  compact finalization success, validation rejection, persistence
  failure, and parse-first notice handling (a complete response
  whose payload text looks like a truncation notice is used
  directly with no saved-file read).
- `.opencode/test/golden-output-shape.test.ts`: compact
  finalization shape at the documented consolidate sites across the
  five consumer commands (both `uf.unleash` steps included),
  summary shape, parsed-key ordering (proposals and warnings before
  findings, correlation before findings), the 15/15/10/5 caps, the
  worst-case all-caps budget, the 16/7 under-40,960-byte budget,
  compact-finalization caps and ordering (10 proposals / 5 warnings,
  both flags, observability before `calculated_assessment` / `errors`),
  and full-mode omission of empty `proposals` / `warnings`.
- `.opencode/test/finalize-error-boundaries.test.ts`: the compact
  wrapper is not a payload — feeding its fields to
  `finalizeReviewDispatch` fails strict root validation, and the
  compact `calculated_assessment` carries only `authoritative`.

### D8. Scaffold mirror and drift gate

Every edited plugin and command file is copied byte-identical into
`internal/scaffold/assets/opencode/`. `TestEmbeddedAssets_MatchSource`
(`internal/scaffold/scaffold_test.go`) enforces the copy.

## Risks / Trade-offs

- **Consumer breakage if a command is missed.** D5 lists all five
  commands; the golden-shape suite covers the review-dispatch
  registry, and the command files are drift-gated. A missed
  consumer reading `review_dispatch` or object `review_verdict`
  would fail at runtime, not at build time, so the task list
  enumerates each command edit explicitly.
- **Spill fallback remains prose.** The summary budget makes it
  exceptional, but a pathological payload could still be truncated
  and the fallback is model-executed. Accepted as an authorized
  DE-001 deviation; D4 records the trigger, owner, and the
  Phase 4 follow-up issue for a narrow read tool.
- **A forged truncation notice could induce a file read.** RD-FR-052
  accepts a path only from the strict host-notice shape, confined to
  the opencode tool-output directory, non-symlink, and at most
  1 MiB, and treats the content as untrusted. The residual risk is a
  file the host itself wrote that happens to match; the command
  never executes or follows spilled text. Complete output that parses
  as JSON is used directly, so notice-shaped payload text cannot
  trigger a read.
- **The size guarantee is by construction, not arithmetic.** The
  summary builder serializes compactly, measures the UTF-8 byte
  length after assembly, and drops trailing entries in a fixed order
  until the result is strictly under 40,960 bytes. A character-count
  derivation would understate escaped or multibyte text; the
  construction keeps the bound unconditional.
- **Summary caps can drop or shorten detail.** A pathological
  dispatch summarizes at most 15 findings and 15 advisories, and
  caps proposals at 10 and warnings at 5. The four truncation flags
  make entry drops explicit, and each included entry carries
  `truncated: true` when any of its own fields was shortened.
  Consumers that need every finding or advisory read the persisted
  artifact at the finalization `artifact_path`; the artifact does
  not carry proposals or warnings. A dropped or shortened proposal
  or warning is **not** recoverable from the RD-FR-052 spill (the
  spill is the same capped summary). That is an accepted
  content-loss trade-off, distinct from the spill prose-execution
  deviation, and the Phase 4 narrow read tool is its remedy.
- **The `runs` array can still outgrow the host window.** Ordering
  only places `runs` ahead of the later arrays; it cannot bound the
  array's own size. Full-mode consumers that consolidate from
  `runs` (`uf.triage-issue.md`, `uf.address-feedback.md`) therefore
  treat a detected truncation notice as a failure signal and fail
  closed instead of consolidating a partial `runs` array. A
  truncated `runs` array whose notice is not detected remains a
  residual risk; the Phase 4 read tool removes it.
- **Size budget can become brittle.** 40 KiB is well above the
  realistic 16-finding summary (roughly 6-8 KiB) and well below
  the observed 63 KiB failure, so it catches regressions without
  flaking on small finding-text changes.
- **`calculated_assessment` shape change.** Callers that read
  `.payload` now get undefined. D5 updates every in-repo caller;
  external consumers must read the artifact path. This is the
  intended reduction in output size.
- **Coverage floor risk.** The new branches (summary builder,
  compact mapper, validation refinement) are small and covered by
  the tests in D7. If the 94 percent floor fails, the gatekeeping
  rule applies: report and stop; never adjust
  `.opencode/vitest.config.ts`.
