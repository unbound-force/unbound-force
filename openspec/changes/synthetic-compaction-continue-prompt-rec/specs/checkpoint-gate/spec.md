<!-- capability: checkpoint-gate
     change: synthetic-compaction-continue-prompt-rec
     originating_issue: 1118
     Mechanical enforcement of checkpoint gate overrides against
     synthetic-prompt quoting, declared-step enforcement, and
     escalated-advisory relay. Design rationale: design.md.
     Companion hardening: harden-spec-review-gate (HARDEN-001..006)
     and compaction-storm-on-a-tight-context-mode (DELEG-FR-001..002).
-->

Requirement numbering: this change allocates SYNTH-001 through
SYNTH-006 for the requirements added below. SYNTH-007 through
SYNTH-020 are reserved for future checkpoint-gate changes.
Archived capability requirements live under
`openspec/specs/<capability>/spec.md`. Requirement ids MUST stay
inside the allocated range.

## ADDED Requirements

### Requirement: Gate overrides carry a human-decision reference

**ID: SYNTH-001** This requirement MUST be satisfied.

Any `checkpoint` call whose `data` object carries a gate
override — any `data` key at any nesting depth containing the
substring `override` (case-insensitive), covering checklist,
spec-review advisory escalation, and quality-gate overrides —
MUST carry a
verifiable human-decision reference: a `human_decision_ref`
naming a `decision_id` present in the persisted human-decision
record (SYNTH-006). `checkpoint` MUST NOT accept an inline
human-decision payload in place of a reference: authorization
always resolves to a pre-existing, machinery-stamped record
entry. Entries reach the record through exactly two surfaces:
the orchestrator relay — `uf_delegate_command` persists entries
from its typed `human_decisions` envelope parameter at spawn
time, before the delegated child session runs — and the
in-session question path — the `record_human_decision` tool
records a `question`-tool answer the session received, under
the controls SYNTH-006 mandates. A free-text justification
alone (for example quoting "Continue if you have next steps, or
stop and ask for clarification if you are unsure how to proceed")
MUST NOT satisfy the requirement. When the reference is absent,
names no recorded entry (a fabricated reference), or otherwise
does not correspond to a recorded human decision, `checkpoint`
MUST reject the call with a machine-parseable error identifying
the missing reference and naming the offending step and
override-carrying data key.

#### Scenario: Synthetic prompt quoted as consent is rejected

- **GIVEN** a child calls `checkpoint` with
  `step: checklist-override-phase8-9` and
  `data.checklist_override` quoting "Continue if you have next
  steps" as "User explicitly directed continuation"
- **AND** no entry in the human-decision record covers that
  override
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST reject with an error indicating a
  missing human-decision reference
- **AND** MUST NOT persist progress state for that step

#### Scenario: Override with a valid human-decision reference is accepted

- **GIVEN** a gate override is requested via `checkpoint` with a
  `human_decision_ref` naming a `decision_id` present in the
  human-decision record for the state key
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST record the step with the reference as
  provenance and return success

#### Scenario: Fabricated reference naming no recorded decision is rejected

- **GIVEN** a child calls `checkpoint` with a `human_decision_ref`
  naming a `decision_id` that is not present in the human-decision
  record for the state key
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST reject the call identically to a
  missing reference, with the `missing_human_decision` error
- **AND** MUST NOT persist progress state for that step

#### Scenario: Override data under a declared step name is rejected

- **GIVEN** a child calls `checkpoint` with the declared step
  `implement` and a `data.checklist_override` field quoting the
  synthetic prompt
- **AND** no `human_decision_ref` is present
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST reject the call with
  `missing_human_decision` naming the step and the
  override-carrying data key
- **AND** MUST NOT persist progress state for that step

#### Scenario: Prior-session free-text override does not carry forward

- **GIVEN** a prior session recorded an override as free text with
  no human-decision reference (e.g. "prior session override for
  Phase 1 remains in effect")
- **WHEN** a new session cites that text as justification without
  its own `human_decision_ref`
- **THEN** `checkpoint` MUST reject the new call as missing a
  human-decision reference

### Requirement: Synthetic prompt is never consent

**ID: SYNTH-002** This requirement MUST be satisfied.

Command prose in `/uf.unleash` and `/uf.run-pipeline-plan` and
the `cobalt-crush-dev` agent prose
(`.opencode/agents/cobalt-crush-dev.md`, the incident child's
persona) MUST state explicitly that compaction prompts and
synthetic continue prompts (including "Continue if you have next
steps, or stop and ask for clarification if you are unsure
how to proceed") are not human input and MUST NOT be treated as
authorization for any gate override. The same prose MUST state
the rejection recovery path at these decision points: a session
whose `checkpoint` call is rejected with `missing_human_decision`,
`undeclared_step`, or `escalated_advisory_unrelayed` MUST stop
and obtain a real human decision via the `question` tool or the
orchestrator relay, and MUST NOT compose a
`record_human_decision` entry for a decision it did not
actually receive. A contract test MUST pin
that prose in both the live copy and the corresponding scaffold
mirror — `.opencode/commands/uf.unleash.md` and
`internal/scaffold/assets/opencode/commands/uf.unleash.md`;
`.opencode/commands/uf.run-pipeline-plan.md` and
`internal/scaffold/assets/opencode/commands/uf.run-pipeline-plan.md`;
`.opencode/agents/cobalt-crush-dev.md` and
`internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md` —
so drift in either copy is a test failure.

#### Scenario: Prose pin prevents regression

- **GIVEN** the synthetic-prompt-not-consent paragraph is present
  in `.opencode/commands/uf.unleash.md`,
  `.opencode/commands/uf.run-pipeline-plan.md`, and
  `.opencode/agents/cobalt-crush-dev.md`, and in each file's
  scaffold mirror
- **WHEN** any copy is edited to remove or alter the paragraph
- **THEN** the contract test MUST fail identifying the missing or
  drifted prose

#### Scenario: Model quoting synthetic prompt is still rejected

- **GIVEN** the prose in SYNTH-002 is present in commands and agent
- **WHEN** a child quotes the synthetic prompt as "User explicitly
  directed continuation" in a `checkpoint` data field
- **THEN** `checkpoint` MUST still reject per SYNTH-001 regardless
  of the quoted text content

### Requirement: Declared checkpoint step enforcement

**ID: SYNTH-003** This requirement MUST be satisfied.

`checkpoint` MUST reject any step name that is not in the
declared step set whenever the call's `command` argument
resolves to `uf.unleash` (the same command keying the
HARDEN-002 spec-review gate uses); `checkpoint` calls under any
other command identifier (`uf.fix-ci`, `uf.review-council`,
`uf.address-feedback`, `uf.finale`, …) carry their own step
vocabularies and MUST be unaffected by this requirement, unless
the human-decision reference required by SYNTH-001 is present
and valid. The declared set for `uf.unleash` is exactly:
`branch-safety-gate`, `tier-routing`,
`resumability-detection`, `clarify`, `plan`, `tasks`,
`spec-review`, `implement`, `implement-phase-N` (N a positive
integer), `code-review`, `quick-pre-flight`, `quick-fix-budget`,
`retrospective`, and `demo` — the union of the pipeline's
user-facing steps, the machinery steps the uf-workflow plugin
auto-checkpoints under command `uf.unleash`, and the gate and
budget steps the `uf.unleash` command prose records. The
allowlist MUST stay in lockstep with both writer surfaces: a
consistency test enumerates the step names prescribed by
`.opencode/commands/uf.unleash.md` checkpoint calls and the
plugin's auto-checkpoint call sites that record under command
`uf.unleash`, and fails when a step name either surface records
under `uf.unleash` is absent from the allowlist. A step name
matching the pattern `*override*` is never in the declared
set, so `checkpoint` MUST reject it whenever the human-decision
reference is absent or invalid.

#### Scenario: Free-form override step name is rejected

- **GIVEN** a child calls `checkpoint` with
  `step: checklist-override-phase8-9`
- **AND** no valid `human_decision_ref` is present
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST reject with an error naming the
  undeclared step and the missing authorization

#### Scenario: Override step with valid human reference is accepted

- **GIVEN** a child calls `checkpoint` with a step matching
  `*override*`
- **AND** a valid `human_decision_ref` per SYNTH-001 is present
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST accept and persist the step

#### Scenario: Normal declared step is unaffected

- **GIVEN** a child calls `checkpoint` with a declared step —
  `implement`, `implement-phase-2`, `branch-safety-gate`,
  `tier-routing`, or `quick-pre-flight`
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST NOT reject the call on step-name
  grounds, with or without a human-decision reference

#### Scenario: Another command's checkpoint is unaffected

- **GIVEN** a session calls `checkpoint` with command
  `uf.review-council` and that command's own step vocabulary
  (for example `phase-1a-preflight`)
- **WHEN** `checkpoint` evaluates the call
- **THEN** `checkpoint` MUST NOT apply the `uf.unleash`
  declared-step predicate to the call

### Requirement: Escalated ADVISORIES are relayed

**ID: SYNTH-004** This requirement MUST be satisfied.

The review-dispatch finalizer's canonical decision maps every
`APPROVE WITH ADVISORIES` consolidation to `ESCALATED`, so
"marked ESCALATED" is not a distinguishing subset of advisory
results. Whenever the latest consolidated spec-review result is
`APPROVE WITH ADVISORIES` (persisted canonical decision
`ESCALATED`), it MUST be relayed to the human via
`relay_spec_review_gate` (introduced in #1094) before
implementation continues, the same relay path used for
`REQUEST CHANGES`. While such an unrelayed result is the latest
consolidated spec-review result under command `uf.unleash` and
no human-decision reference per SYNTH-001 covers it,
`checkpoint` MUST reject the advancement steps `implement` and
`implement-phase-N` (any positive integer N) with
`escalated_advisory_unrelayed`, mirroring the HARDEN-002 hold
shape (a hold over advancement steps, not an allowlist of
permitted ones). A plain `APPROVE` consolidation (canonical
decision `APPROVED`) does not require the relay.

#### Scenario: Escalated advisories are relayed before implementation

- **GIVEN** spec-review consolidates to
  `APPROVE WITH ADVISORIES (ESCALATED)` with
  `correlation 8ef4f276-…`, 16 findings and 2 proposals
- **AND** no human-decision reference covers the escalation
- **WHEN** the pipeline attempts to advance from `spec-review` to
  `implement`
- **THEN** the pipeline MUST call `relay_spec_review_gate` and
  present the relay to the human
- **AND** MUST NOT advance to `implement` until the relayed
  decision is recorded

#### Scenario: Plain APPROVE does not require the relay

- **GIVEN** spec-review consolidates to `APPROVE` (persisted
  canonical decision `APPROVED`)
- **WHEN** the pipeline evaluates advancement
- **THEN** advancement MAY proceed without the SYNTH-004 relay

### Requirement: Regression replay of the incident checkpoint

**ID: SYNTH-005** This requirement MUST be satisfied.

The repository MUST include a deterministic regression test that
replays the exact `checkpoint` JSON from
`ses_e200afc75ffeSuftvVXtQnBJhR` (command `uf.unleash`, step
`checklist-override-phase8-9`, data
`checklist_override: "User explicitly directed continuation via
'Continue if you have next steps'…"`, no `human_decision_ref`)
and asserts `checkpoint` rejects the call. The test MUST also
assert that the synthetic prompt text alone never satisfies
SYNTH-001.

#### Scenario: Regression fixture replays incident JSON

- **GIVEN** the exact incident checkpoint payload is supplied as
  a fixture (step `checklist-override-phase8-9` with the
  quoted synthetic prompt in `checklist_override`)
- **WHEN** the test invokes `checkpoint` with that payload
- **THEN** the test MUST observe a rejection per SYNTH-001 and
  SYNTH-003
- **AND** MUST NOT observe persisted progress for that step

### Requirement: Human-decision record store contract

**ID: SYNTH-006** This requirement MUST be satisfied.

The human-decision record
(`.uf/progress/<key>/human-decisions.json`) MUST be an
append-only store whose entries carry a machine-generated
`decision_id` (UUID v4), the source (`question-tool` or
`orchestrator-relay`), the question and answer text, and a
machine-generated `recorded_at`; callers supply only source,
question, and answer. The record path MUST be derived through
the same sanitized key resolution as the progress state store —
slug sanitization, path-traversal rejection, symlink refusal,
directory mode 0o700 and file mode 0o600 — and the record MUST
remain excluded from version control (it lives under the
gitignored `.uf/progress/` tree). `record_human_decision` MUST
refuse entries whose answer, after normalization (trimming,
whitespace collapsing, case folding), contains the normalized
synthetic compaction prompt text as a substring. Writes MUST be
atomic (temporary file plus rename), and appends MUST be
serialized — an advisory lock file or equivalent
exclusive-write protocol — so concurrent
`record_human_decision` calls cannot lose entries. An absent
record file is the empty-record case: a gate override citing it
is rejected with `missing_human_decision` because no entry can
resolve. Record reads that fail otherwise — permission denied,
corrupted JSON, or any other non-ENOENT I/O error — MUST fail
closed: `checkpoint` rejects any gate override citing the
unreadable record with a machine-parseable
`human_decision_record_unreadable` error rather than treating
the failure as a missing reference or crashing. Input domains
are bounded: `human_decision_ref` at most 256 characters and
matching the UUID v4 pattern of the `decision_id` it names,
step names at most 128 characters, question and answer at most
16 KiB each; violations MUST be rejected with a machine-parseable
error before any security-sensitive operation (file path
derivation, error construction, state persistence). Errors and
logs MUST name bounded identifiers (`decision_id`, step, data
key) — never raw question or answer text — and question or
answer content MUST NOT be logged at INFO or above.

#### Scenario: Traversal-shaped key is rejected

- **GIVEN** a caller supplies a state key or reference shaped to
  escape `.uf/progress/` (for example `../../etc`)
- **WHEN** the record path is derived
- **THEN** the derivation MUST reject the value with a
  machine-parseable error
- **AND** no file outside `.uf/progress/` is read or written

#### Scenario: Unreadable record fails closed

- **GIVEN** the human-decision record for the state key is
  corrupted or unreadable
- **WHEN** `checkpoint` evaluates a gate override citing the
  record
- **THEN** `checkpoint` MUST reject with
  `human_decision_record_unreadable`
- **AND** MUST NOT persist progress state for that step

#### Scenario: First-use missing record is the empty-record case

- **GIVEN** no human-decision record exists yet for the state
  key (first use)
- **WHEN** `checkpoint` evaluates a gate override citing a
  `human_decision_ref`
- **THEN** `checkpoint` MUST reject with `missing_human_decision`
  (no entry can resolve), not
  `human_decision_record_unreadable`

#### Scenario: Synthetic prompt answer is refused at recording time

- **GIVEN** a caller invokes `record_human_decision` with an
  answer matching the synthetic compaction prompt up to
  normalization — case, whitespace, or quoting variation, or
  embedded inside a larger justification
- **WHEN** the tool evaluates the entry
- **THEN** the tool MUST refuse to append the entry and return a
  machine-parseable error

## MODIFIED Requirements

None. This change does not modify existing requirements; all
requirements are additions to the checkpoint-gate capability.

## REMOVED Requirements

None. This change removes no requirements.
