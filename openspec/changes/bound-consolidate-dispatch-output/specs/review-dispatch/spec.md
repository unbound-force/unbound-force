<!-- capability: review-dispatch
     change: bound-consolidate-dispatch-output
     originating_issue: 922
     Transport-safety contract for consolidate_dispatch: bounded
     retrieval, order-stable full output, compact finalization,
     and a deterministic retrieval-recovery path. Design rationale
     and rejected alternatives: design.md (D1-D8). -->

Requirement numbering: RD-FR-049 through RD-FR-053 are allocated
for this change. RD-FR-054 through RD-FR-058 are reserved for
future review-dispatch changes. The RD-FR-001..048 range is
allocated across the review-dispatch capability's unarchived and
archived deltas (RD-FR-001..008 review-council-multi-model-fanout,
RD-FR-025..027 fix-review-dispatch-recovery, RD-FR-028..031
derive-per-run-timeout, RD-FR-032..036
review-dispatch-diff-manifest, RD-FR-037 host-fallback-provenance,
RD-FR-038 split-host-resolution-failure-classes, with RD-FR-039..048
reserved there); the review-matrix capability carries
RD-FR-009..010 and RD-FR-019..024, with RD-FR-011..018 reserved
there. The unnumbered `dispatch_agent_run Tool` requirement lives
in the dispatch-agent-run capability delta (shared-agent-executor,
as MODIFIED by host-fallback-provenance); it is not changed here.
Requirement ids MUST stay inside the allocated range.

Archive order: the sibling change `dispatch-finalize-determinism` MUST
archive before this change, or in the same archive batch. Its
`dispatch-finalization` FR-001 finalizer-validity guarantee and its
review-council requirement "Finalize Every Iteration Via
Consolidation" both describe the finalizer's returned data. RD-FR-051
is the reciprocal reference: it scopes the FR-001 guarantee to
`detail: "full"` and defines the compact finalization result as the
authoritative outcome and scalar decision, with the full projection
read once from `artifact_path`. OpenSpec cannot carry MODIFIED deltas
for either requirement here because neither target spec exists in
`openspec/specs/` until that change archives.

## ADDED Requirements

### Requirement: Bounded Consolidation Retrieval

**ID: RD-FR-049** This requirement MUST be satisfied.

`consolidate_dispatch` MUST accept an optional `detail` input whose
value is either `"full"` or `"summary"` and whose default is
`"full"`. The input MUST remain closed-validated: any other value
MUST be rejected as a validation failure before any run file is
read.

When `detail` is `"summary"` and `provenance` is omitted, the tool
MUST return exactly the following top-level keys, in this insertion
order: `correlation_id`, `proposals`, `warnings`, `verdict`,
`verdict_reason`, `run_counts`, `coverage`, `findings`, `advisories`,
`findings_truncated`, `advisories_truncated`, `proposals_truncated`,
and `warnings_truncated`. The optional keys are omitted when not
applicable: `proposals` and `warnings` when their respective arrays
are empty (mirroring the existing full-mode behavior), and the
truncation flags when no entry was dropped. Placing the correlation
identifier and the observability arrays before the findings keeps
them recoverable when a host truncates the tail of the output. The
summary result MUST NOT carry `artifact_path`, `plan`, `runs`,
`input_context`, `change_profile`, `plan_version`, `workflow_result`,
`command`, `mode`, or `full`: no artifact is persisted by a
retrieval-only call, so an artifact path has no source.

The full mode MUST keep its existing schema-exact payload and
observability fields for backward compatibility. The summary mode
is a retrieval-only contract: supplying `detail: "summary"`
together with `provenance` MUST be rejected as a validation
failure rather than silently ignored or applied to finalization.

The serialized summary result for a dispatch with 16 unique
findings and 7 collected lesson proposals MUST be strictly smaller
than 40 KiB (40,960 bytes) when the retrieval is
`consolidate_dispatch` with `detail: "summary"` and the result is
serialized with compact (no-indentation) JSON. This budget is the
regression guard for host-side truncation; it MUST be asserted by
the plugin test suite.

The fixture budget alone does not bound a pathological dispatch: a
session may carry up to 512 findings, 512 advisories, and a
result-level proposals array whose entries may each reach 16,384
characters. The summary builder MUST therefore apply runtime caps and
a slim entry projection to every array it returns.

A summary finding MUST carry only `severity` (the existing enum),
`category` (at most 128 characters), `description` (at most 512
characters), `file` (at most 256 characters or null), and `line` (an
integer or null); `root_cause` and `run_ids` MUST NOT be emitted in
summary mode. A summary advisory MUST carry only `severity` and
`description` (at most 512 characters); its `run_ids` MUST NOT be
emitted. The builder MUST include at most 15 findings and at most 15
advisories. A summary proposal MUST carry only `agent` (at most 128
characters), `information` (at most 512 characters), `tag` (at most
128 characters), and the optional `category`; the builder MUST
include at most 10 proposals. A summary warning MUST be projected to
`kind` (one of the five documented warning kinds, at most 64
characters) and `detail` (at most 256 characters), with `detail`
composed from the structured warning fields in a fixed order per
kind: `orphaned_submissions_excluded` becomes
`agent=<agent> run=<run_id> findings=<n> proposals=<n>`;
`invalid_submissions_excluded` becomes
`agent=<agent> run=<run_id> invalid=<n>`; `run_findings_overflow`
becomes `agent=<agent> kept=<n> dropped=<n>`;
`consolidated_findings_overflow` becomes `kept=<n> dropped=<n>`;
and `empty_coverage_attestation` becomes `agent=<agent>`. The
builder MUST include at most 5 warnings. The summary
`verdict_reason` MUST be truncated to at most 512 characters.

Each included entry MUST carry `truncated: true` when any of its own
emitted fields was shortened; the marker MUST be omitted otherwise.
When a cap drops entries the result MUST set the corresponding
`findings_truncated`, `advisories_truncated`, `proposals_truncated`,
or `warnings_truncated` flag to `true`; when no entry is dropped the
flag MUST be omitted.

The caps MUST NOT alter the persisted artifact: the complete
`payload.findings` and `payload.advisories` remain at the compact
finalization result's `artifact_path`, and consumers that need the
full arrays MUST read them there. Proposals and warnings are
result-level data that the finalizer payload does not persist, so an
entry dropped by these caps is NOT recoverable from the RD-FR-052
spill (the spill is the same capped summary). That accepted
content-loss trade-off is recorded in the design and remedied only by
the follow-up narrow read tool. Host-side truncation of the tool
output is a different loss mode and remains spill-recoverable.

The bound MUST hold by construction, not by arithmetic. The summary
builder MUST serialize the summary with compact (no-indentation)
JSON and measure its UTF-8 byte length after assembly. While that
length is 40,960 bytes or greater, the builder MUST drop entries in
a fixed order — trailing findings first, then trailing advisories,
then trailing proposals, then trailing warnings — setting the
corresponding truncation flag for each array that loses an entry.
The loop terminates because a summary with every array empty is far
below the budget: `verdict_reason` is capped at 512 characters and
every scalar field is bounded. A schema-legal or multibyte-heavy
dispatch therefore still yields a result strictly smaller than
40 KiB (40,960 bytes). The test suite MUST assert the construction
with multibyte, escape-heavy, and every-field-at-maximum fixtures in
addition to the 16-finding/7-proposal fixture.

#### Scenario: Summary retrieval excludes the dispatch body

- **GIVEN** a persisted dispatch session with two successful runs
- **AND** `consolidate_dispatch` is called without provenance and
  with `detail: "summary"`
- **WHEN** the call completes
- **THEN** the returned object carries `correlation_id`, `verdict`,
  `verdict_reason`, `run_counts`, `coverage`, `findings`, and
  `advisories`
- **AND** it carries no `plan`, `runs`, `input_context`,
  `change_profile`, `plan_version`, `workflow_result`, `command`,
  `mode`, or `full` key

#### Scenario: Summary observability precedes the findings

- **GIVEN** a summary retrieval whose result carries findings,
  advisories, at least one proposal, and at least one warning
- **WHEN** the tool serializes the result
- **THEN** the key order of the parsed object places `proposals`,
  `warnings`, and `correlation_id` before `findings`
- **AND** the ordering assertion reads `Object.keys(JSON.parse(s))`
  rather than substring offsets, so field names appearing inside
  finding, proposal, or verdict text cannot collide with it

#### Scenario: Runtime caps bound a pathological summary

- **GIVEN** a dispatch session whose runs carry 512 unique findings,
  512 advisories, 512 lesson proposals, and 512 warnings
- **WHEN** `consolidate_dispatch` returns `detail: "summary"`
- **THEN** the result carries at most 15 findings and at most 15
  advisories
- **AND** every emitted finding carries only `severity`, `category`,
  `description`, `file`, and `line` — never `root_cause` or
  `run_ids` — with `description` at most 512 characters and `file`
  at most 256 characters
- **AND** every emitted advisory carries only `severity` and
  `description` — never `run_ids`
- **AND** the result carries at most 10 proposals with `information`
  at most 512 characters and `tag` at most 128 characters
- **AND** the result carries at most 5 warnings projected to `kind`
  and `detail`
- **AND** all four truncation flags are `true`
- **AND** no artifact is persisted by this retrieval-only call

#### Scenario: Worst-case caps stay inside the budget

- **GIVEN** a summary result in which every array is at its cap
  (15 findings, 15 advisories, 10 proposals, 5 warnings) and every
  included field is at its maximum: `category` and `tag` 128
  characters, finding and advisory `description` 512, `file` 256,
  proposal `information` 512, warning `kind` 64, warning `detail`
  256, `verdict_reason` 512
- **WHEN** the tool serializes the result
- **THEN** the serialized output is strictly smaller than 40,960
  bytes and the test asserts the construction bound

#### Scenario: Default remains full

- **GIVEN** `consolidate_dispatch` is called without provenance and
  without a `detail` field
- **WHEN** the call completes
- **THEN** the returned payload keeps the legacy full shape with
  `plan`, `runs`, `input_context`, `change_profile`,
  `plan_version`, and `workflow_result`
- **AND** existing consumers that omit `detail` continue to work

#### Scenario: Unknown detail value is rejected

- **GIVEN** `consolidate_dispatch` is called with
  `detail: "everything"`
- **WHEN** the input is validated
- **THEN** the call returns a validation failure
- **AND** no dispatch session directory is read

#### Scenario: Summary mode cannot finalize

- **GIVEN** `consolidate_dispatch` is called with both
  `detail: "summary"` and a `provenance` object
- **WHEN** the input is validated
- **THEN** the call returns a validation failure
- **AND** no artifact is persisted

#### Scenario: Sixteen findings and seven proposals stay in budget

- **GIVEN** a successful dispatch session whose runs carry 16
  unique findings and 7 lesson proposals
- **WHEN** `consolidate_dispatch` returns `detail: "summary"`
- **THEN** the serialized output is smaller than 40,960 bytes
- **AND** all 7 proposals are present in the output

### Requirement: Retrieval Payload Ordering

**ID: RD-FR-050** This requirement MUST be satisfied.

The full retrieval result (no provenance) MUST serialize its keys in
this insertion order: `command`, `mode`, `full`, `input_context`,
`change_profile`, `plan_version`, `correlation_id`, `proposals`,
`warnings`, `verdict`, `verdict_reason`, `run_counts`, `coverage`,
`runs`, `plan`, `findings`, `advisories`, and `workflow_result`.
`proposals` and `warnings` MUST be omitted when their arrays are
empty. The scalar decision fields MUST precede the large arrays, and
`runs` MUST precede `plan`, `findings`, and `advisories` so that a
tail truncation consumes those later arrays before the `runs` array
that `uf.triage-issue.md` and `uf.address-feedback.md` depend on.
The ordering change MUST NOT alter any value, only insertion order
of the emitted object.

The full-mode result is consumed by commands whose consolidation
rules read `runs`; the reordering exists to protect both the
observability arrays and that consumer cargo. Ordering MUST be
asserted on the key order of the parsed object
(`Object.keys(JSON.parse(s))`), not on substring offsets.

The `proposals` and `warnings` keys MUST remain result-level
observability fields. They MUST NOT be added to the payload handed
to `finalizeReviewDispatch`, whose schema is strict at the root.

#### Scenario: Proposals precede the dispatch body

- **GIVEN** a dispatch session with at least one collected lesson
  proposal and at least one warning
- **WHEN** `consolidate_dispatch` returns the full result without
  provenance
- **THEN** the parsed object's key order places `proposals`,
  `warnings`, and `correlation_id` before `plan` and before `runs`
- **AND** `runs` precedes `plan`, `findings`, and `advisories`

#### Scenario: Empty observability keys are omitted

- **GIVEN** a dispatch session with no collected proposals and no
  warnings
- **WHEN** `consolidate_dispatch` returns the full result without
  provenance
- **THEN** the output carries neither a `proposals` nor a
  `warnings` key

#### Scenario: Finalizer payload stays strict

- **GIVEN** a provenance-bearing `consolidate_dispatch` call
- **WHEN** the tool builds the payload for
  `finalizeReviewDispatch`
- **THEN** the payload contains no `proposals` or `warnings` key
- **AND** finalization succeeds or fails on the existing strict
  schema without wrapper-field rejection

### Requirement: Compact Finalization Response

**ID: RD-FR-051** This requirement MUST be satisfied.

When `consolidate_dispatch` finalizes internally (provenance
supplied), its returned object MUST carry these top-level keys in
this insertion order: `status`, `operation_verdict`,
`correlation_id`, `artifact_path`, `review_verdict`, `run_counts`,
then the optional observability keys `proposals`, `warnings`,
`proposals_truncated`, and `warnings_truncated`, then
`calculated_assessment`, then `errors`. Optional keys are omitted
when empty or false. The observability keys MUST precede
`calculated_assessment` and `errors` so a host-truncated tail
cannot drop them.

`proposals` and `warnings` in this result are capped and projected
exactly as RD-FR-049 defines for the summary: at most 10 proposals
(each `agent` at most 128 characters, `information` at most 512,
`tag` at most 128, optional `category`, plus `truncated: true`
when any emitted field was shortened) and at most 5 warnings (each
`kind` at most 64 characters, `detail` at most 256, plus the same
marker). `proposals_truncated` and `warnings_truncated` are true
only when entries were dropped. This cap is an accepted risk
recorded beside DE-001: the persisted artifact carries findings
and advisories but not proposals or warnings, so a dropped or
shortened proposal or warning is not recoverable from that
artifact; the Phase-4 narrow read tool is the remedy.

`review_verdict` MUST be the scalar canonical decision string
(`"APPROVED"`, `"CHANGES_REQUESTED"`, `"ESCALATED"`,
`"INCONCLUSIVE"`, or `"UNAVAILABLE"`) when a projected review
verdict exists, and null otherwise. `run_counts` MUST be the
payload's run-count object when a payload is available, and null
otherwise.

`calculated_assessment` MUST be either null or an object carrying
only the scalar `authoritative` boolean. It MUST NOT carry
`payload` or `review_verdict` members. The returned object MUST NOT
carry `review_dispatch` or any other copy of the persisted artifact
body.

The compact result is authoritative for the finalization outcome
and for the scalar decision. The full projection object
(`review_dispatch`, including `payload.findings` and
`payload.advisories`) remains the persisted artifact, and consumers
that need it MUST read it once from `artifact_path`. A consumer MUST
NOT describe the compact result as carrying the projection.

This requirement scopes the dispatch-finalization FR-001
finalizer-validity guarantee to retrieval with `detail: "full"`
(the default): a full-mode response remains a payload
`finalize_review_dispatch` accepts unmodified, while a
`detail: "summary"` response is a reporting projection only and
MUST NOT be submitted to the finalizer. Finalization always uses
the provenance-bearing `consolidate_dispatch` call
(dispatch-finalization FR-002).

The `authoritative` flag is the structured discriminator the
command uses for bounded mechanical retry: null means the payload
was rejected at validation and no artifact was persisted; false
means validation passed but persistence failed; true means the
artifact was persisted. Consumers MUST be able to classify the
failure from the structured response alone, without parsing error
prose.

#### Scenario: Summary responses are not finalizer-valid

- **GIVEN** a `consolidate_dispatch` retrieval with
  `detail: "summary"`
- **WHEN** its returned object is submitted to
  `finalize_review_dispatch` unmodified
- **THEN** validation rejects it because `plan` and `runs` are
  absent
- **AND** callers must use the provenance-bearing finalization
  call instead

#### Scenario: Successful finalization is compact

- **GIVEN** a valid provenance-bearing finalization that persists
  an artifact
- **WHEN** `consolidate_dispatch` returns
- **THEN** the result carries `status: "success"`,
  `operation_verdict`, `correlation_id`, and `artifact_path`
- **AND** `review_verdict` is the scalar council decision string
- **AND** `run_counts` equals the payload's run counts
- **AND** `calculated_assessment` is `{ "authoritative": true }`
- **AND** the parsed key order is `status`, `operation_verdict`,
  `correlation_id`, `artifact_path`, `review_verdict`,
  `run_counts`, any observability keys, `calculated_assessment`,
  `errors`
- **AND** the result carries no `review_dispatch` key and no
  payload body

#### Scenario: Finalization caps observability without losing order

- **GIVEN** a provenance-bearing finalization whose dispatch
  produced 12 proposals and 7 warnings
- **WHEN** `consolidate_dispatch` returns the compact result
- **THEN** `proposals` carries 10 entries and `warnings` carries
  5 entries
- **AND** `proposals_truncated` and `warnings_truncated` are true
- **AND** both keys precede `calculated_assessment` and `errors`
- **AND** the persisted artifact at `artifact_path` still carries
  the complete findings and advisories arrays

#### Scenario: Validation rejection is distinguishable

- **GIVEN** a provenance-bearing finalization whose payload fails
  structural or semantic validation
- **WHEN** `consolidate_dispatch` returns
- **THEN** `status` is `"failed"` and `artifact_path` is null
- **AND** `errors` enumerates the validation defects
- **AND** `calculated_assessment` is null

#### Scenario: Persistence failure is distinguishable

- **GIVEN** a valid payload whose artifact write fails
- **WHEN** `consolidate_dispatch` returns
- **THEN** `status` is `"failed"` and `artifact_path` is null
- **AND** `calculated_assessment` is
  `{ "authoritative": false }`
- **AND** the result carries no payload body

### Requirement: Deterministic Retrieval Recovery

**ID: RD-FR-052** This requirement MUST be satisfied.

The `uf.review-council` command's Section 4 retrieval MUST call
`consolidate_dispatch` without provenance, with the captured
`correlation_id`, and with `detail: "summary"`. It MUST NOT
instruct the agent to "use this single payload and do not
re-fetch it" without a bounded recovery path.

The host truncation notice is observed to be two lines: the
first reports that the tool output was truncated and carries
the absolute path of a saved file holding the complete
untruncated result; the second recommends an explore-delegation
recovery. Observed saved files are complete, pretty-printed
JSON objects of the tool result — the iteration-2 spec-review
result (22 findings) was saved as a roughly 234 KB file under
the opencode tool-output directory, while the truncated output
was roughly 65 KB. The notice's second line MUST be ignored:
the command reads the saved file itself exactly once.

Before scanning for a notice, the command MUST attempt to parse the
entire retrieval output as a single JSON object. If that parse
succeeds, the output is complete: the command MUST use it directly
and MUST ignore any notice-shaped text inside it.

When the parse fails and no line matches the truncation-notice shape
below, the command MUST treat the retrieval as unrecoverable. When a
line matches, the command MUST treat the saved output path as
untrusted data and MUST accept it only when both conditions hold:

1. At least one line of the output matches the case-insensitive
   regular expression
   `^\[?[^\n]*\btruncat\w*\b[^\n]*\bsaved\b[^\n]*[:=]\s*("[^"]+"|\/[^\s\]]+)\]?$`
   and yields exactly one path capture. A double-quoted capture
   MAY contain spaces (legal home and XDG paths can); the
   surrounding quotes are stripped before validation. When more
   than one line matches, the command MUST use the single line
   whose captured path satisfies condition 2; if no line or more
   than one line satisfies it, the notice is ambiguous and MUST
   be rejected.
2. The captured path is absolute; contains no `..` path
   segment; resolves to a regular file whose realpath starts
   with the opencode tool-output directory's realpath followed
   by a `/` separator (strictly beneath the directory, never the
   directory itself) (`$XDG_DATA_HOME/opencode/tool-output`
   when `XDG_DATA_HOME` is set, otherwise
   `~/.local/share/opencode/tool-output`); has no symbolic link
   at any path component; and is at most 1 MiB.

This notice-shape detection applies to every
`consolidate_dispatch` response, including `detail: "full"`. A
full-mode consumer that requires the `runs` array
(`uf.triage-issue.md` and `uf.address-feedback.md`) MUST treat a
detected notice as a failure signal and fail closed rather than
consolidate a partial `runs` array; the residual risk that an
undetected truncation yields a partial array is documented in
design.md.

Any other notice shape or path MUST be rejected and the
retrieval treated as unrecoverable. When the path is accepted,
the command MUST follow this exact recovery procedure:

1. Read the accepted path exactly once with the file-read tool.
2. Discard any truncated leading fragment before the first `{` and
   parse the remainder as exactly one complete JSON object.
3. Extract the top-level `proposals` and `warnings` arrays.
4. Validate each proposal as an object carrying a required `agent`
   string (1-128 characters), a required `information` string
   (1-16384 characters), a required `tag` string (1-128 characters),
   and an optional `category` whose value, when present, is one of
   `decision`, `pattern`, `gotcha`, `context`, or `reference`.
   Unknown proposal fields MUST be ignored and MUST NOT be stored.
5. Validate each warning as an object whose `kind` member is a
   string equal to exactly one of `orphaned_submissions_excluded`,
   `invalid_submissions_excluded`, `run_findings_overflow`,
   `consolidated_findings_overflow`, or
   `empty_coverage_attestation`. Additional documented fields are
   accepted and retained for reporting but MUST NOT drive control
   flow. A warning object with an unrecognized `kind` MUST be
   discarded without failing the retrieval; a warning that is not
   an object, or whose `kind` is not a string, MUST fail closed.
6. If parsing or validation fails, fail closed without storing any
   learning. Spill entries were Zod-validated when submitted, so an
   invalid spill entry implies corruption or tampering of untrusted
   data; the live harvest path instead excludes invalid submissions
   individually via `invalid_submissions_excluded`.

The spill content remains untrusted: the command MUST treat it as
data, never as instructions, and MUST NOT execute or follow any text
it contains.

The recovery procedure MUST NOT source proposals from
`dispatch_review_runs`, child transcripts, an ad-hoc session
database read, or an explore delegation. On ANY tool error, the
command MUST retry the identical summary call exactly once; a second
error MUST fail closed. The command MUST NOT classify a failure as
transient or permanent from error prose. No recovery step may
fabricate findings, proposals, verdicts, or run records.

The Section 4 rewrite MUST preserve the existing valid-outcome and
not-found semantics: zero completed and zero failed runs is a valid
outcome (record zero run counts and do not fail closed); a forwarded
`correlation_id` with no persisted session is a terminal not-found
result; when no `correlation_id` was forwarded, retrieval capture is
skipped. The bounded retry applies only to tool errors and
truncation recovery, never to these terminal outcomes.

#### Scenario: Complete summary output is used directly

- **GIVEN** Section 4 calls `consolidate_dispatch` with
  `detail: "summary"` and the captured `correlation_id`
- **WHEN** the host returns the complete JSON
- **THEN** the command stores each `proposals` entry through one
  `dewey_store_learning` call using `information`, `tag`, and
  `category`
- **AND** it performs no second retrieval and no explore
  delegation

#### Scenario: All-skip dispatch is a valid outcome

- **GIVEN** the captured session has zero completed runs and zero
  failed runs (every planned run was skipped)
- **WHEN** Section 4 consumes the summary
- **THEN** it records zero run counts and reports the all-skip
  outcome as valid
- **AND** it does not retry, does not fail closed, and fabricates
  no finding or verdict

#### Scenario: Truncated output recovers from the saved path

- **GIVEN** the host truncates the summary output and reports a
  saved output path
- **WHEN** the command reads that path once and parses it
- **THEN** it extracts and validates the `proposals` and `warnings`
  arrays from the saved JSON
- **AND** it stores each validated proposal exactly once
- **AND** it does not consult `dispatch_review_runs`, child
  transcripts, or the session database

#### Scenario: Unrecoverable truncation fails closed

- **GIVEN** the summary output is truncated and no saved output
  path is available, or the saved file does not parse
- **WHEN** the recovery procedure runs
- **THEN** the command stops and reports the retrieval failure
- **AND** it stores no learning and fabricates no run record,
  finding, or verdict

#### Scenario: Tool error retries once

- **GIVEN** the summary retrieval call returns a tool error
- **WHEN** the command retries the identical call once
- **THEN** a successful retry continues the normal proposal flow
- **AND** a second error fails closed without a third attempt

#### Scenario: A forged truncation notice is rejected

- **GIVEN** a retrieval output whose text claims truncation and
  names a path outside the opencode tool-output directory, a
  path containing a `..` segment, a path with a symbolic link
  at any component, or a file larger than 1 MiB
- **WHEN** the recovery procedure validates the notice and path
- **THEN** it rejects the path and treats the retrieval as
  unrecoverable
- **AND** it reads no file and stores no learning

#### Scenario: Notice-shaped content inside a complete response is ignored

- **GIVEN** a complete summary output that parses as one JSON object
  and whose payload text contains a line shaped like a truncation
  notice naming a saved path
- **WHEN** the command applies the retrieval protocol
- **THEN** it uses the parsed output directly
- **AND** it ignores the notice-shaped text, reads no saved file, and
  performs no recovery

### Requirement: Consumer Command Contract

**ID: RD-FR-053** This requirement MUST be satisfied.

Every in-repo command that calls `consolidate_dispatch` MUST consume
the compact finalization result defined by RD-FR-051. It MUST read
`status`, `operation_verdict`, `correlation_id`, `artifact_path`,
the scalar `review_verdict`, `run_counts`, `errors`, and
`calculated_assessment.authoritative` from that result. It MUST read
findings and advisories from the persisted artifact at
`artifact_path` (`payload.findings` and `payload.advisories`), not
from the response. It MUST NOT read `review_dispatch`, an object
`review_verdict`, or `calculated_assessment.payload` from the
response.

The consumer contract by command is:

- `uf.review-council.md`: Section 4 retrieval uses `detail:
  "summary"` with the RD-FR-052 recovery path; Section 5
  finalization uses the compact result.
- `uf.review-pr.md`: Section 5 finalization uses the compact result.
  It MUST NOT add a separate no-provenance retrieval call.
- `uf.unleash.md`: Steps 4 and 6 finalize with provenance, consume
  the compact result, and read findings and advisories once from
  `artifact_path`.
- `uf.triage-issue.md`: Section 2.3 retrieval remains `detail:
  "full"` because its three-rule verdict consolidation reads the
  `runs` array; Section 4.6 finalization uses the compact result.
- `uf.address-feedback.md`: Section 2.3.7 retrieval remains `detail:
  "full"`; Section 2.3.8 finalization uses the compact result.

A command that needs proposals or warnings MUST source them from the
compact finalization result when it finalized, or from a summary
retrieval only where this requirement names one. No command may
reconstruct proposals from run files, child transcripts, or the
session database.

#### Scenario: Finalization consumers read the artifact for details

- **GIVEN** a command finalizes a dispatch with provenance
- **WHEN** it consumes the compact result
- **THEN** it reads `review_verdict` as a scalar string and
  `calculated_assessment.authoritative` as a boolean
- **AND** it reads findings and advisories from
  `artifact_path`'s `payload.findings` and `payload.advisories`
- **AND** it reads no `review_dispatch`, object `review_verdict`, or
  `calculated_assessment.payload`

#### Scenario: Triage and feedback keep full retrieval

- **GIVEN** `uf.triage-issue.md` Section 2.3 or
  `uf.address-feedback.md` Section 2.3.7 retrieves a dispatch
- **WHEN** the command calls `consolidate_dispatch` without
  provenance
- **THEN** it uses `detail: "full"`
- **AND** its verdict or feedback consolidation still sees the
  `runs` array

#### Scenario: Review-pr adds no retrieval call

- **GIVEN** `uf.review-pr.md` finalizes a dispatch with provenance
- **WHEN** it consumes the compact result
- **THEN** it obtains proposals and warnings from that result
- **AND** it performs no separate no-provenance retrieval call

## REMOVED Requirements

None.
