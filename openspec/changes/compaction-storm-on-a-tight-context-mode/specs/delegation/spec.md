<!-- capability: delegation
     change: compaction-storm-on-a-tight-context-mode
     originating_issue: 1079
     rationale: new capability covering selection headroom
     gating, compaction-storm detection, and the final-text
     success guard; design decisions D1-D4 live in this
     change's design.md
-->

Requirement numbering: this change allocates DELEG-FR-001
through DELEG-FR-003 for the requirements added below.
DELEG-FR-004 through DELEG-FR-018 are reserved for future
delegation changes. The progress-accounting contract this
change builds on lives in the sibling capability
`delegation-timeouts` (DT-FR-001 through DT-FR-004); this
change modifies DT-FR-002 in a separate delta under
specs/delegation-timeouts/ in this same change directory.
The selection precedence chain and audit trail live in the
sibling capability `model-matrix` (MM-FR-003, MM-FR-007;
delta in openspec/changes/model-matrix-delegation/).
Requirement ids MUST stay inside the allocated range.

## ADDED Requirements

### Requirement: Selection headroom gate

**ID: DELEG-FR-001** This requirement MUST be satisfied.

A delegation MUST NOT be dispatched to a model whose
declared usable input window is smaller than the
delegation's base footprint plus a working margin. The
usable window is the declared context limit minus the
declared output reserve; the base footprint is the prompt
material the delegation necessarily carries (cached prefix
plus command body), supplied by the caller as part of the
selection inputs so that selection itself performs no
filesystem access. When a candidate's usable window is
insufficient, selection MUST reject or rank down the
candidate and MUST record the reason in the selection
fallback reason, so the audit trail explains why a
preferred model was not used (the MM-FR-007 surface).

The caller-side footprint estimator MUST constrain its
reads to repository-root command paths resolved by
path-joining; it MUST reject paths that traverse outside
the repository root or resolve through symlinks to
locations outside it, and MUST NOT read arbitrary paths.
On any read failure it MUST fall back to a documented
conservative default footprint sized so that a read
failure cannot understate the footprint, and MUST record
the fallback in the selection audit.

A candidate that declares no usable window MUST NOT be
rejected on headroom grounds; selection MUST proceed to
the next precedence check. A candidate whose declared
limits are internally inconsistent (an output reserve at
least as large as the context window, leaving no usable
input) MUST be rejected with a reason naming the
inconsistency.

When every candidate is rejected for headroom, selection
MUST terminate deterministically: the delegation MUST fall
back to the host model (or the configured fallback) and
the fallback reason MUST record the full rejection chain.
Selection MUST NOT silently proceed with a rejected
candidate and MUST NOT stall without a result.

Concrete knob names and defaults (for example a working
margin of 24k tokens in the initial implementation) are
non-normative; the normative content is the reject,
record, and terminate behavior above. When the working
margin is configurable, it MUST be validated on load: a
value of the wrong type or outside the documented range
MUST be replaced by the documented default, and the
substitution MUST be logged as a warning naming the key,
the invalid value, and the default applied. A
misconfigured margin (for example zero, or a value that
would admit any candidate) MUST NOT silently disable the
headroom gate.

#### Scenario: Tight-context classifier pick falls back

- **GIVEN** an orchestrator delegation whose base
  footprint is roughly 76.5k tokens of cached prefix plus
  the command body, with a working margin configured
- **AND** the classifier-preferred model declares limits
  leaving roughly 98k usable input
- **WHEN** the delegation is dispatched
- **THEN** the child runs on a model with sufficient
  headroom (the next eligible candidate or the host
  fallback), not on the classifier-preferred model
- **AND** the selection fallback reason names the rejected
  model, its usable window, the required window, the
  footprint, the margin, and the fallback taken

#### Scenario: All candidates rejected falls back to host

- **GIVEN** a delegation where every candidate model is
  rejected for headroom or inconsistent limits
- **WHEN** selection completes
- **THEN** the delegation proceeds on the host model (or
  the configured fallback)
- **AND** the fallback reason records each rejected
  candidate and its rejection reason
- **AND** no rejected candidate is silently selected

#### Scenario: No declared limits is not a rejection

- **GIVEN** a candidate model with no declared context or
  output limits
- **WHEN** headroom gating runs
- **THEN** the candidate is not rejected for headroom and
  selection proceeds to the next precedence check

#### Scenario: Inconsistent limits rejected with reason

- **GIVEN** a candidate model declaring an output reserve
  at least as large as its context window
- **WHEN** headroom gating runs
- **THEN** the candidate is rejected with a fallback
  reason naming the inconsistent declaration

#### Scenario: Estimator guards path escapes and read failures

- **GIVEN** a footprint estimation request whose command
  path contains traversal segments, resolves through a
  symlink to a location outside the repository root, or
  names a file that cannot be read
- **WHEN** the estimator computes the footprint
- **THEN** the path is rejected or the read failure falls
  back to the documented conservative default footprint
- **AND** the rejection or fallback is recorded in the
  selection audit
- **AND** the resulting footprint never understates the
  delegation's prompt material on a read failure

#### Scenario: Misconfigured margin falls back to default

- **GIVEN** a configuration that sets the working margin
  to an invalid value (for example zero, or a non-integer)
- **WHEN** the configuration is loaded
- **THEN** the margin is replaced by the documented
  default
- **AND** a warning naming the key, the invalid value,
  and the default applied is logged
- **AND** the headroom gate still rejects candidates
  whose usable window cannot fit the footprint plus the
  default margin

### Requirement: Compaction-storm detection

**ID: DELEG-FR-002** This requirement MUST be satisfied.

The delegation executor MUST detect a compaction storm --
a run of consecutive context-compaction events with no
intervening substantive progress -- and MUST abort the
child with a retryable failure carrying durable resume
state, well before wall-clock exhaustion. A delegated
child storming in a compaction loop on a tight-context
model terminates within minutes with a retryable failure
and a resume envelope, not after hours of success-masked
thrashing.

Compaction events (context-compression summaries and
synthetic continuation prompts injected into the session)
MUST NOT count as progress for idle-timeout or storm
accounting; the accounting is shared with the idle
watchdog (DT-FR-002). Substantive progress is the
successful completion of a non-read tool call, a file
write, or a checkpoint record (a non-exhaustive example
list). Read-only tool calls (file reads, searches, and
other inspection) MUST NOT count as substantive progress
whatever their outcome, and a non-read tool call that
fails or has not completed does not count: only
successful completions reset the storm window.

When the configured number of consecutive compaction
events accrues without intervening substantive progress,
the executor MUST abort the child and return a failure
that is retryable, carries a code distinct from the idle
and wall-clock timeout codes, names the storm window in
its error details, and includes a resume envelope with
durable state (the project, agent, and last recorded
checkpoint step, reusing the existing resume envelope
shape). The threshold is configurable with a documented
default (the initial implementation uses 3, a
non-normative example). The threshold MUST be validated
on load: a value of the wrong type or outside the
documented range MUST be replaced by the documented
default, and the substitution MUST be logged as a warning
naming the key, the invalid value, and the default
applied. A misconfigured threshold (for example one no
run could realistically reach) MUST NOT silently disable
the storm detector.

#### Scenario: Storm loop ends in minutes, not hours

- **GIVEN** a delegated child that enters a compaction
  loop on a tight-context model (roughly one compaction
  every 30 seconds with no substantive progress between
  them)
- **WHEN** the configured storm threshold is reached
- **THEN** the delegation ends within minutes with a
  retryable failure and a resume envelope carrying durable
  state
- **AND** the failure is not reported as success

#### Scenario: Threshold reached triggers abort

- **GIVEN** a running child where the probe observes the
  configured number of consecutive compaction markers with
  zero substantive progress between them
- **WHEN** the threshold-reaching compaction is recorded
- **THEN** the executor aborts the child and returns a
  retryable failure with a code distinct from idle and
  wall-clock timeouts
- **AND** a resume envelope carries the project, agent,
  and pending checkpoint step

#### Scenario: Substantive progress resets the window

- **GIVEN** a child one compaction short of the storm
  threshold since its last substantive progress
- **WHEN** substantive progress occurs (the successful
  completion of a non-read tool call, a file write, or a
  checkpoint record) before the next compaction
- **THEN** the compaction counter resets and no storm
  failure is emitted until the threshold again accrues
  without progress

#### Scenario: Compaction is not probe progress

- **GIVEN** a child whose recent subtree activity consists
  solely of compaction summaries
- **WHEN** progress accounting runs for idle-timeout
  purposes
- **THEN** those messages do not count as progress and the
  idle watchdog may fire independently of storm detection

#### Scenario: Reads do not reset the window

- **GIVEN** a child one compaction short of the storm
  threshold since its last substantive progress
- **WHEN** read-only tool calls complete between the
  remaining compactions, with no successful non-read tool
  call, file write, or checkpoint record
- **THEN** the compaction counter does not reset and the
  storm threshold triggers

#### Scenario: Misconfigured threshold falls back to default

- **GIVEN** a configuration that sets the storm threshold
  to an invalid value (for example out of the documented
  range)
- **WHEN** the configuration is loaded
- **THEN** the threshold is replaced by the documented
  default
- **AND** a warning naming the key, the invalid value,
  and the default applied is logged
- **AND** a compaction loop still aborts through the
  storm detector rather than running for hours

### Requirement: Final-text success guard

**ID: DELEG-FR-003** This requirement MUST be satisfied.

A child whose final message text is a compaction summary
or a synthetic continuation or handoff prompt MUST NOT be
reported as success. The executor MUST reclassify such a
result as a retryable failure: the storm code when the
storm threshold was met, otherwise a distinct
synthetic-final-text code. The original text MUST be
preserved in the error for debugging in sanitized form.
The sanitization guarantees are normative: the preserved
text MUST be truncated to a documented maximum length,
MUST NOT carry cached-prefix content (system prompt
material, earlier tool output, or other context that
preceded the final message), and MUST have secrets
redacted through the existing error-sanitization path. A
result that stores raw, unbounded, or prefix-carrying
final text does not satisfy this requirement.

The classifier MUST match structurally: anchored,
case-insensitive, and length-bounded matchers that
recognize the synthetic wrapper structure (for example a
compaction-summary wrapper or a known synthetic prompt
prefix), never bare substrings that can appear in
legitimate text. A legitimate final message MUST remain
success.

#### Scenario: Compaction-summary final text is not success

- **GIVEN** a child that exits with final text that
  structurally matches a compaction summary and recorded
  no checkpoint
- **WHEN** the executor maps the result to a status
- **THEN** the result is not success; it is a retryable
  failure with the storm or synthetic-final-text code
- **AND** the preserved original text in the error is
  truncated to the documented maximum, carries no
  cached-prefix content or secrets, and is otherwise
  sanitized

#### Scenario: Synthetic handoff prompt is not success

- **GIVEN** a child whose final text is the injected
  synthetic continuation prompt
- **WHEN** the executor classifies the final text
- **THEN** the result is reclassified to a retryable
  failure with the synthetic-final-text code (or the storm
  code when the threshold is also met) and no success
  envelope is emitted

#### Scenario: Legitimate final text remains success

- **GIVEN** a child whose final text is a normal
  completion or gate-pause message, including one that
  merely mentions compaction in prose without matching
  the structural wrappers
- **WHEN** the executor classifies the final text
- **THEN** the existing success mapping is preserved

## MODIFIED Requirements

None. The delegation capability is new; no existing
requirement is modified. The compaction-exclusion change
to idle-timeout progress accounting modifies DT-FR-002 in
the delegation-timeouts capability (delta under
specs/delegation-timeouts/ in this change directory).

## REMOVED Requirements

None.
