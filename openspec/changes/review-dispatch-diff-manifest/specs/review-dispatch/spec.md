<!-- capability: review-dispatch
     change: review-dispatch-diff-manifest
     originating_issue: 877 -->

## ADDED Requirements

### Requirement: Diff Delivery Mode Selection

**ID: RD-FR-032** This requirement MUST be satisfied.

The `build_review_prompts` tool MUST select a diff delivery mode
from the diff's line count for every non-triage mode invocation:
inline mode at or below `DIFF_THRESHOLD` (1500 lines, value
unchanged), manifest mode strictly above it. The selection MUST
be a deterministic function of the diff content: identical diff
content MUST select the identical mode. The threshold constant
and the line-counting rule that feeds it MUST have a single
shared definition in the new manifest lib module — one
definition of both the constant and the counting function —
which the plugin re-exports or imports; the constant's value
MUST remain 1500 lines. Every consumer of the threshold
(prompt-assembly mode selection and the capture-path
`exceeds_threshold` computation) MUST use the shared counting
function, so the capture-path warning fires if and only if
manifest mode engages.

At or below the threshold, the assembled prompt for each agent
MUST be byte-for-byte identical to the pre-change inline
assembly for the same inputs — zero behavior change, including
the `## Diff` section, untrusted delimiters, and the 2 MiB
`MAX_DIFF_BYTES` inline read cap.

In manifest mode the diff content MUST NOT be inlined into any
prompt. The tool MUST still read the diff file tool-side to
construct the manifest, bounded by the existing capture-side
write caps (8 MiB captured diff; PR-diff persistence path).
`MAX_PROMPT_BYTES` (4 MiB) MUST remain enforced on the
assembled prompt in both modes. Triage mode (issue content, no
diff) MUST be unaffected by this requirement.

Where the capture path previously warned that the diff exceeds
the advisory threshold, the warning MUST state that manifest
delivery is engaged for the dispatch rather than implying
degraded quality. The capture tool whose behavior this
modifies (`fetch_pr_review_context`) is declared as a
modified capability with its own delta under
specs/fetch-pr-review-context/, whose FR-006/FR-010
restatements preserve the pinned "Diff exceeds size advisory
threshold" warning substring.

#### Scenario: Below threshold stays inline

- **GIVEN** a non-triage `build_review_prompts` invocation whose
  diff contains exactly 1500 lines
- **WHEN** prompts are assembled
- **THEN** every agent prompt contains the full diff in the
  `## Diff` untrusted section
- **AND** each prompt is byte-for-byte identical to the
  pre-change assembly for the same inputs

#### Scenario: Above threshold engages manifest mode

- **GIVEN** a non-triage `build_review_prompts` invocation whose
  diff contains 1501 lines
- **WHEN** prompts are assembled
- **THEN** no agent prompt contains the inlined diff content
- **AND** every agent prompt contains the `## Diff Manifest`
  section per RD-FR-035

#### Scenario: Deterministic mode selection

- **GIVEN** the same diff content
- **WHEN** mode selection runs twice
- **THEN** both invocations select the identical delivery mode

#### Scenario: Cross-tool counting agreement at the boundary

- **GIVEN** diffs containing exactly 1500 and exactly 1501
  lines
- **WHEN** the capture path computes `exceeds_threshold` and
  prompt assembly selects the delivery mode
- **THEN** both computations use the shared counting function
- **AND** the capture-path warning fires if and only if
  prompt assembly engages manifest mode

#### Scenario: Triage mode unaffected

- **GIVEN** a triage-mode invocation carrying `issue_content`
- **WHEN** prompts are assembled
- **THEN** the delivery-mode logic does not alter the triage
  prompt assembly in any way

#### Scenario: Prompt size cap still enforced

- **GIVEN** a manifest-mode invocation whose assembled prompt
  would exceed 4 MiB
- **WHEN** prompts are assembled
- **THEN** the tool fails with `prompt_size_exceeded`

### Requirement: Diff Manifest Construction

**ID: RD-FR-033** This requirement MUST be satisfied.

A manifest module (`.opencode/lib/review-diff-manifest.ts`)
MUST construct a diff manifest as a pure function of the patch
content: identical patch content MUST produce an identical
manifest (stable file ordering following patch order; stable
chunk identifiers derived from file path and chunk index).

The manifest MUST carry one entry per file section in the
patch, including the synthesized new-file sections that
`captureReviewDiff` emits for untracked and git-ignored files.
Each entry MUST include: the file path, addition and deletion
counts, binary and rename flags where the patch declares them,
and the list of retrievable chunks for that file. Each chunk
MUST carry an identifier and coordinates into the patch file
(byte offsets), and chunk boundaries MUST be hunk-aligned:
boundaries MUST fall on `@@` hunk headers or file-section
boundaries. A single hunk larger than the chunk byte budget MAY
be split mid-hunk as a last resort, and such chunks MUST be
annotated as mid-hunk splits.

Each chunk MUST NOT exceed the chunk byte budget (256 KiB). Each
entry MUST carry persona-relevance hints derived from
deterministic path-pattern classification aligned with the
`reviewer-capabilities.yaml` scope names (security,
dependencies, standard, cli-ux, ci-cd, documentation,
test-quality). Hints are guidance only: the manifest MUST list
every changed file for every agent, and hints MUST NOT filter,
omit, or reorder entries out of patch order.

When `build_review_prompts` operates in manifest mode it MUST
persist the manifest as a machine-readable JSON artifact with
restrictive file permissions (0o600) resolvable from the diff
path, recording the source diff path, the diff's byte length,
and the manifest schema version. The manifest MUST record a
SHA-256 digest of the diff content for integrity verification
at retrieval time.

#### Scenario: Deterministic manifest

- **GIVEN** the same patch content
- **WHEN** the manifest is constructed twice
- **THEN** both constructions produce identical JSON

#### Scenario: Hunk-aligned chunk boundaries

- **GIVEN** a file section of the patch spanning multiple
  hunks and exceeding the 256 KiB chunk budget
- **WHEN** chunks are computed
- **THEN** every chunk boundary falls on an `@@` hunk header
  or the file-section boundary
- **AND** no chunk exceeds 256 KiB

#### Scenario: Oversized single hunk splits annotated

- **GIVEN** a single hunk larger than the chunk byte budget
- **WHEN** chunks are computed
- **THEN** the hunk is split into chunks each within the
  budget where possible
- **AND** each resulting chunk is annotated as a mid-hunk split

#### Scenario: Synthesized new-file sections included

- **GIVEN** a captured diff containing a synthesized
  `diff --git` section for an untracked file
- **WHEN** the manifest is constructed
- **THEN** the manifest carries an entry for that file with
  its addition count and retrievable chunks

#### Scenario: Hints never filter

- **GIVEN** a patch mixing workflow files, documentation, test
  files, and Go sources
- **WHEN** the manifest is constructed
- **THEN** every file appears in the manifest in patch order
- **AND** each entry carries at least one scope hint
- **AND** no entry is omitted or reordered on the basis of its
  hint

### Requirement: Bounded Diff Chunk Retrieval Tool

**ID: RD-FR-034** This requirement MUST be satisfied.

The review-dispatch plugin MUST register a `read_diff_chunk`
tool that serves diff content to reviewer agents on demand. The
tool input MUST accept a diff path and either a chunk identifier
or a file path, and the tool MUST validate: the diff path
against the existing path-validation rules; the presence and
loadability of the manifest persisted for that diff; and the
existence of the requested chunk or file in that manifest. Every
validation failure MUST return a structured fail-closed error.

The tool MUST be confined to the manifest: it MUST NOT read
files other than the path-validated diff recorded in the
manifest, and it MUST NOT return content outside
manifest-declared chunk coordinates. The reviewer
CONFINEMENT_RULE scope is unchanged by this tool.

The tool MUST verify manifest-to-diff integrity before serving
(recorded byte length and SHA-256 digest) and MUST fail
closed on mismatch. The tool's response MUST be bounded to at
most 256 KiB of diff content per call; a file section larger
than the budget MUST return its first chunk plus the
identifiers of the remaining chunks. Served content MUST be
wrapped in untrusted-evidence delimiters with the existing
cross-section delimiter escaping and a defensive preamble
treating the content as data — the same injection-safety
guarantee as inline delivery.

#### Scenario: Chunk served within budget

- **GIVEN** a manifest-mode dispatch and a chunk identifier
  present in its manifest
- **WHEN** `read_diff_chunk` is called with the diff path and
  that chunk identifier
- **THEN** the tool returns the chunk's diff content
- **AND** the content does not exceed 256 KiB
- **AND** the content is enclosed in untrusted-evidence
  delimiters

#### Scenario: File selector returns first chunk and
continuations

- **GIVEN** a manifest entry whose file section spans multiple
  chunks
- **WHEN** `read_diff_chunk` is called with the diff path and
  that file path
- **THEN** the tool returns the file's first chunk
- **AND** identifies the remaining chunk identifiers for that
  file

#### Scenario: Unknown chunk fails closed

- **GIVEN** a chunk identifier absent from the manifest
- **WHEN** `read_diff_chunk` is called
- **THEN** the tool returns a structured error
- **AND** no diff content is returned

#### Scenario: Path outside validation rejected

- **GIVEN** a diff path that fails the existing path-validation
  rules
- **WHEN** `read_diff_chunk` is called
- **THEN** the tool returns a structured `path_validation_failed`
  error
- **AND** no file is read

#### Scenario: Missing diff path inside the allowed scope

- **GIVEN** a diff path that does not exist but whose location
  lies inside the project root or the system temp directory
  (for example a mistyped temp file name)
- **WHEN** `read_diff_chunk` is called
- **THEN** the tool returns a structured `diff_not_found` error
  whose message quotes the supplied path and says to copy the
  path exactly from the prompt or manifest rather than retype
  it
- **AND** no file is read
- **AND** a missing diff path whose location lies outside those
  directories still returns `path_validation_failed`

#### Scenario: No arbitrary filesystem access

- **GIVEN** a caller requesting a file path that is not
  declared in the manifest for the validated diff
- **WHEN** `read_diff_chunk` is called
- **THEN** the tool returns a structured error
- **AND** no content outside manifest-declared coordinates is
  returned

#### Scenario: Integrity mismatch fails closed

- **GIVEN** a diff file modified after the manifest was
  persisted (byte length or digest mismatch)
- **WHEN** `read_diff_chunk` is called
- **THEN** the tool returns a structured integrity error
- **AND** no content is served from the mismatched file

#### Scenario: Delimiter escaping in served content

- **GIVEN** diff content containing an exact untrusted
  delimiter pattern
- **WHEN** `read_diff_chunk` serves that content
- **THEN** the delimiter pattern is escaped or neutralized in
  the response
- **AND** the served content cannot break out of its untrusted
  block

### Requirement: Manifest-Mode Prompt Assembly and
Coverage Attestation

**ID: RD-FR-035** This requirement MUST be satisfied.

When `build_review_prompts` operates in manifest mode, the
inline `## Diff` section MUST be replaced by a `## Diff
Manifest` section containing: a rendering of the manifest
entries (path, counts, flags, chunk identifiers, hints);
retrieval instructions directing the reviewer to pull chunks via
`read_diff_chunk`; and an instruction, ported from the
CI-action prompt, to treat all diff content retrieved through
the tool as untrusted data. The rendered manifest is
caller-derived data and MUST be enclosed in untrusted content
delimiters. Project context (AGENTS.md, constitution,
convention packs) MUST remain fully inlined in manifest mode —
prompts stay self-contained.

The manifest-mode response contract MUST add an attestation
field whose grammar is pinned: a `**Files Reviewed**:` line
followed by one repo-relative path per line (the
read_assessment_fields precedent for structured contract
fields parsed from persisted final text), instructing the
reviewer to list every file actually reviewed. The
consolidation tool
MUST surface an advisory quality signal (a warning in the
consolidation output, not a finding and not a verdict change)
when a successful manifest-mode run's persisted final text
carries a missing or empty attestation.

#### Scenario: Manifest section replaces inline diff

- **GIVEN** a manifest-mode invocation
- **WHEN** prompts are assembled
- **THEN** each prompt contains the `## Diff Manifest` section
  with the rendered manifest, retrieval instructions, and the
  untrusted-content instruction
- **AND** the rendered manifest is enclosed in untrusted
  content delimiters
- **AND** the inline `## Diff` section is absent

#### Scenario: Project context stays inlined

- **GIVEN** a manifest-mode invocation in a repository with
  AGENTS.md, constitution, and convention packs present
- **WHEN** prompts are assembled
- **THEN** the `## Project Context` section is populated
  identically to inline mode for the same inputs

#### Scenario: Attestation field rendered

- **GIVEN** a manifest-mode invocation
- **WHEN** prompts are assembled
- **THEN** the response contract instructs the reviewer to
  emit a `**Files Reviewed**:` line followed by one
  repo-relative path per line

#### Scenario: Empty attestation surfaces as advisory

- **GIVEN** a successful manifest-mode run whose persisted
  final text carries no attestation or an empty one
- **WHEN** the dispatch is consolidated
- **THEN** the consolidation output carries an advisory
  warning identifying the run
- **AND** the verdict derivation and finding set are unchanged
  by the advisory

#### Scenario: Populated attestation produces no advisory

- **GIVEN** a successful manifest-mode run whose persisted
  final text carries a non-empty attestation
- **WHEN** the dispatch is consolidated
- **THEN** no empty-attestation advisory is emitted for that
  run

### Requirement: Retrieval Allowance in Timeout Derivation

**ID: RD-FR-036** This requirement MUST be satisfied.

The per-run timeout derivation (RD-FR-028/RD-FR-029) MUST gain
a named additive retrieval allowance applied when — and only
when — the run's prompt was assembled in manifest mode. The
delivery-mode signal MUST flow deterministically from prompt
assembly to run dispatch, including the recovery path: the
signal MUST be persisted among the run's persisted work
signals (or re-derived from the persisted prompt), so
`redispatch_failed_runs` — which derives each re-run's window
from persisted work signals rather than re-running prompt
assembly — grants the same retrieval allowance to a
re-dispatched manifest-mode run. The allowance MUST be a fixed
constant
recorded in the derivation's timeout provenance record as an
optional field, following the run-file-only audit record
pattern: older run files without the field MUST remain valid.

The allowance MUST NOT alter any other derivation semantics:
the composition order, coefficients, contention multiplier,
floor, ceiling, and the explicit-argument override MUST behave
exactly as RD-FR-028/RD-FR-029 specify — with RD-FR-028's
composition restated by this change's MODIFIED delta to
include the retrieval allowance as its fifth term. The
derivation MUST remain a pure function of its inputs.

#### Scenario: Manifest-mode run gains the allowance

- **GIVEN** two otherwise-identical dispatches, one assembling
  prompts in manifest mode and one in inline mode
- **WHEN** per-run timeouts are derived with no explicit
  argument
- **THEN** the manifest-mode run's effective timeout is greater
  than the inline-mode run's by exactly the retrieval allowance
  (absent clamping)
- **AND** the manifest-mode run's provenance records the
  allowance

#### Scenario: Inline-mode runs unchanged

- **GIVEN** an inline-mode dispatch
- **WHEN** per-run timeouts are derived
- **THEN** the derivation is identical to the pre-change
  derivation for the same inputs
- **AND** no retrieval allowance term is applied

#### Scenario: Explicit argument still wins

- **GIVEN** a manifest-mode run dispatched with an explicit
  per-run timeout argument
- **WHEN** the timeout resolves
- **THEN** the explicit value is the effective timeout
- **AND** the allowance does not modify it

#### Scenario: Older run files remain valid

- **GIVEN** run files persisted before this requirement
- **WHEN** the dispatch tooling reads them
- **THEN** the files parse and consolidate without requiring
  the retrieval-allowance provenance field

#### Scenario: Redispatched manifest-mode run retains the
allowance

- **GIVEN** a dispatch session holding a failed manifest-mode
  run whose delivery-mode signal was persisted among its work
  signals
- **WHEN** `redispatch_failed_runs` re-runs that reviewer
  without a timeout argument
- **THEN** the re-run's derivation grants the manifest-mode
  retrieval allowance
- **AND** the re-run's effective timeout exceeds an
  otherwise-identical inline-mode re-run's by exactly the
  allowance (absent clamping)

## MODIFIED Requirements

### Requirement: Per-Run Timeout Derivation

**ID: RD-FR-028** This requirement MUST be satisfied.

The dispatch tools MUST derive each reviewer run's timeout
from the work being dispatched via a deterministic formula in
the tool layer, whenever no explicit per-run timeout argument
is provided. The derivation MUST be a pure function of its
inputs: identical inputs MUST produce an identical timeout.

The derivation MUST compose, in order:

1. a base window resolved per RD-FR-029 (matrix
   `limits.per_run_timeout_seconds`, or the applicable
   per-tier override);
2. a prompt-size term that is monotonically non-decreasing in
   the run's prompt size in bytes;
3. a tier-based change-profile term that is monotonically
   non-decreasing in the run's resolved tier (lightweight,
   standard, heavy; zero for tier-less direct-model runs) —
   change-profile size is carried by the prompt-size term;
4. a contention multiplier strictly greater than 1, applied
   when the dispatch permits more than one concurrent run
   (the configured `max_parallel_runs` exceeds 1);
5. a manifest-mode retrieval allowance applied when — and
   only when — the run's prompt was assembled in manifest
   delivery mode (RD-FR-036).

The derived value MUST be clamped to a configurable ceiling
that MUST NOT exceed 1_800_000 milliseconds (the existing
`TimeoutSchema`/`validateTimeout` bound). The derivation MUST
NOT reduce the effective timeout below the base window: the
configured limit remains a floor. When the configured ceiling
is below the resolved base window, the base floor MUST take
precedence: the effective ceiling is the base window. An
explicit per-run timeout argument MUST override the derived
value entirely.

A redispatch re-run executes sequentially; its derivation
MUST NOT apply the contention multiplier. A redispatched
manifest-mode run MUST retain the retrieval allowance: the
delivery-mode signal is persisted among the run's work
signals (RD-FR-036), so the recovery-path derivation grants
the same allowance as the fresh-dispatch path.

Previously: the composition was a closed four-term list
(base window; prompt-size term; change-profile term;
contention multiplier) and the redispatch clause said nothing
about delivery mode. This change adds the fifth term — the
manifest-mode retrieval allowance, specified in full by
RD-FR-036 — and extends the redispatch clause so recovered
manifest-mode runs keep it.

#### Scenario: Small work, no contention

- **GIVEN** a dispatch with a small prompt, a lightweight
  change profile, and `max_parallel_runs` of 1
- **WHEN** the per-run timeout is derived with no explicit
  argument
- **THEN** the effective timeout equals the base window
- **AND** no prompt, profile, or contention terms increase it

#### Scenario: Larger prompts scale the window

- **GIVEN** two dispatches identical except that one carries
  a materially larger prompt for the same reviewer
- **WHEN** per-run timeouts are derived for both
- **THEN** the larger prompt's effective timeout is greater
  than or equal to the smaller prompt's

#### Scenario: Parallel contention multiplies the window

- **GIVEN** a dispatch whose `max_parallel_runs` exceeds 1
- **WHEN** the per-run timeout is derived
- **THEN** the effective timeout is greater than or equal to
  the same derivation with `max_parallel_runs` of 1
- **AND** the unclamped derivation is strictly greater than
  the same derivation with `max_parallel_runs` of 1

#### Scenario: Derived value clamped at the ceiling

- **GIVEN** work signals large enough that the unclamped
  derivation exceeds the configurable ceiling
- **WHEN** the per-run timeout is derived
- **THEN** the effective timeout equals the ceiling
- **AND** the ceiling does not exceed 1_800_000 milliseconds

#### Scenario: Explicit argument bypasses derivation

- **GIVEN** any combination of work signals
- **WHEN** the caller provides an explicit per-run timeout
  argument
- **THEN** the explicit value is the effective timeout
- **AND** no derivation term modifies it

#### Scenario: Determinism

- **GIVEN** the same base window, prompt size, change
  profile, concurrency, delivery mode, and ceiling
- **WHEN** the derivation runs twice
- **THEN** both invocations produce the identical effective
  timeout
