<!-- capability: delegation
     change: compaction-storm-detector-sums-compactio
     originating_issue: 1125
     rationale: refines DELEG-FR-002 (Compaction-storm
     detection) so storms are counted per session, executor-
     managed descendants never trip the parent's detector, and
     the storm failure names the storming session; design
     decisions live in this change's design.md
-->

Requirement numbering: this change modifies DELEG-FR-002
(Compaction-storm detection); no new delegation ids are
allocated, and the DELEG-FR-004 through DELEG-FR-018 reservation
from openspec/changes/compaction-storm-on-a-tight-context-mode/
is unchanged. The shared progress-accounting contract this
requirement builds on lives in the sibling capability
`delegation-timeouts` (DT-FR-002); this change refines the
shared-accounting clause in a separate delta under
specs/delegation-timeouts/ in this same change directory.
Requirement ids MUST stay inside the allocated range.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Compaction-storm detection

**ID: DELEG-FR-002** This requirement MUST be satisfied.

The delegation executor MUST detect a compaction storm -- a run
of consecutive context-compaction events with no intervening
substantive progress -- and MUST abort the child with a retryable
failure carrying durable resume state, well before wall-clock
exhaustion. A delegated child storming in a compaction loop on a
tight-context model terminates within minutes with a retryable
failure and a resume envelope, not after hours of
success-masked thrashing.

Compaction events (context-compression summaries and synthetic
continuation prompts injected into the session) MUST NOT count
as progress for idle-timeout or storm accounting; the accounting
is shared with the idle watchdog (DT-FR-002). Substantive
progress is the successful completion of a non-read tool call, a
file write, or a checkpoint record (a non-exhaustive example
list). Read-only tool calls (file reads, searches, and other
inspection) MUST NOT count as substantive progress whatever
their outcome, and a non-read tool call that fails or has not
completed does not count: only successful completions reset a
storm window.

Storm accounting MUST be scoped per session. Consecutive
compaction events accrue against the session in which they
occur, and a session's run resets only on substantive progress
in that same session. Compaction events observed in different
sessions MUST NOT be summed: when no single counted session has
reached the configured number of consecutive compactions, the
detector MUST NOT abort, even when the subtree total across
sessions reaches or exceeds that number. Substantive progress in
one session MUST NOT reset another session's run.

A descendant session that is itself an executor-managed run -- a
session for which a delegation executor has registered a run
deadline, and which therefore carries its own always-armed storm
detector and wall clock -- MUST NOT trip the parent's storm
detector: its compaction markers MUST NOT accrue toward the
parent's per-session counts. A storming executor-managed
descendant ends its own run through its own detector while the
parent delegation continues. The probed root session -- the
delegation's direct child -- is always subject to storm counting
regardless of its own registered deadline. When a descendant
cannot be recognized as executor-managed (for example a session
created in another host process, where the deadline registry is
not shared), its compactions MUST still be counted per session
under the per-session rule above, never summed.

When the configured number of consecutive compaction events
accrues without intervening substantive progress in one counted
session, the executor MUST abort the child and return a failure
that is retryable, carries a code distinct from the idle and
wall-clock timeout codes, names the storm window and the
storming session in its error details, and includes a resume
envelope with durable state (the project, agent, and last
recorded checkpoint step, reusing the existing resume envelope
shape). The limit details of a storm failure MUST identify the
storming session in machine-parseable form, so recovery can
distinguish one thrashing session from a slow, healthy
dispatch. The threshold is configurable with a documented
default (the initial implementation uses 3, a non-normative
example). The threshold MUST be validated on load: a value of
the wrong type or outside the documented range MUST be replaced
by the documented default, and the substitution MUST be logged
as a warning naming the key, the invalid value, and the default
applied. A misconfigured threshold (for example one no run
could realistically reach) MUST NOT silently disable the storm
detector.

#### Scenario: Storm loop ends in minutes, not hours

- **GIVEN** a delegated child that enters a compaction loop on
  a tight-context model (roughly one compaction every 30
  seconds with no substantive progress between them)
- **WHEN** the configured storm threshold is reached in the
  child's own session
- **THEN** the delegation ends within minutes with a retryable
  failure and a resume envelope carrying durable state
- **AND** the failure is not reported as success

#### Scenario: Threshold reached triggers abort

- **GIVEN** a running child where the probe observes the
  configured number of consecutive compaction markers in one
  counted session with zero substantive progress between them
- **WHEN** the threshold-reaching compaction is recorded
- **THEN** the executor aborts the child and returns a
  retryable failure with a code distinct from idle and
  wall-clock timeouts
- **AND** a resume envelope carries the project, agent, and
  pending checkpoint step
- **AND** the limit details identify the storming session

#### Scenario: Substantive progress resets that session's window

- **GIVEN** a session one compaction short of the storm
  threshold since its last substantive progress
- **WHEN** substantive progress occurs in that session (the
  successful completion of a non-read tool call, a file write,
  or a checkpoint record) before its next compaction
- **THEN** that session's compaction counter resets and no
  storm failure is emitted until the threshold again accrues
  without progress in a counted session
- **AND** the counters of other sessions are unaffected

#### Scenario: Compaction is not probe progress

- **GIVEN** a child whose recent subtree activity consists
  solely of compaction summaries
- **WHEN** progress accounting runs for idle-timeout purposes
- **THEN** those messages do not count as progress and the
  idle watchdog may fire independently of storm detection

#### Scenario: Reads do not reset the window

- **GIVEN** a session one compaction short of the storm
  threshold since its last substantive progress
- **WHEN** read-only tool calls complete between the remaining
  compactions in that session, with no successful non-read
  tool call, file write, or checkpoint record
- **THEN** that session's compaction counter does not reset and
  the storm threshold triggers

#### Scenario: Sibling sessions do not sum

- **GIVEN** a delegation whose subtree contains three sibling
  reviewer sessions, each with one new compaction marker and
  only completed read-only tool calls in between
- **WHEN** the probe folds the subtree with a storm threshold
  of three
- **THEN** no session's per-session count reaches the
  threshold and the delegation is not aborted
- **AND** the delegation continues running with in-flight
  children intact

#### Scenario: Executor-managed descendant does not trip the parent

- **GIVEN** a delegation whose subtree contains a descendant
  session for which another executor has registered a run
  deadline (a dispatched reviewer or a nested delegation with
  its own executor and limits)
- **WHEN** that descendant accrues the configured number of
  consecutive compactions without its own substantive progress
- **THEN** the parent delegation is not aborted by the parent's
  storm detector
- **AND** the storming descendant's own run ends through its
  own storm detector as a retryable failure, which its caller
  (for example a dispatch) handles while the delegation
  continues

#### Scenario: Unrecognized descendant still counted per session

- **GIVEN** a delegation whose subtree contains a descendant
  session that cannot be recognized as executor-managed (no
  registered deadline is visible to the probe)
- **WHEN** that descendant alone accrues the configured number
  of consecutive compactions without substantive progress
- **THEN** the parent's storm detector aborts the delegation
  with the storm failure naming that descendant

#### Scenario: Storm envelope names the storming session

- **GIVEN** a delegation aborted by the storm detector because
  one counted session reached the threshold
- **WHEN** the failure is returned
- **THEN** the error message names the storming session and its
  consecutive compaction count
- **AND** the limit details carry the storming session's id in
  machine-parseable form alongside the storm kind

#### Scenario: Misconfigured threshold falls back to default

- **GIVEN** a configuration that sets the storm threshold to an
  invalid value (for example out of the documented range)
- **WHEN** the configuration is loaded
- **THEN** the threshold is replaced by the documented default
- **AND** a warning naming the key, the invalid value, and the
  default applied is logged
- **AND** a compaction loop still aborts through the storm
  detector rather than running for hours

Previously: DELEG-FR-002 counted consecutive compactions
against a single subtree-wide accumulator -- markers in any
session of the probed subtree incremented one
`compactionsSinceProgress` counter, and substantive progress in
any session reset it. Parallel read-heavy children (spec-review
reviewers on a tight-context model) each compacting fewer than
the threshold therefore summed past it and aborted healthy
delegations with `child_active: true` (issue #1125: four false
storm aborts on 2026-11-15). This delta keeps the requirement's
abort envelope, threshold validation, and #1079 protection
(the probed root stays counted) while scoping the counting per
session and excluding executor-managed descendants, whose own
always-armed detectors own their storms.

## REMOVED Requirements

None.
