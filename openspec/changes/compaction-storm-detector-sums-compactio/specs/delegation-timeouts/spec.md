<!-- capability: delegation-timeouts
     change: compaction-storm-detector-sums-compactio
     originating_issue: 1125
     rationale: refines the DT-FR-002 shared-accounting clause so
     compaction accounting is per session; the idle clock itself
     is unchanged and the per-session storm semantics live in the
     delegation capability (DELEG-FR-002)
-->

Requirement numbering: this change modifies DT-FR-002 (Idle
Timeout); no new delegation-timeouts ids are allocated, and the
DT-FR-005 through DT-FR-012 reservation from
openspec/changes/delegation-timeout-budgets/ is unchanged. Under
OpenSpec replacement semantics this MODIFIED delta replaces the
requirement's prose and scenarios; the predecessor scenarios are
carried forward unchanged except for the shared-accounting
clause, which this change scopes per session (issue #1125).
Requirement ids MUST stay inside the allocated range.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Idle Timeout

**ID: DT-FR-002** This requirement MUST be satisfied.

When an idle timeout is configured, the executor MUST poll the
child's session subtree for activity and MUST abort the child
with error code `idle_timeout` (retryable) only after no
activity change has been observed for the idle window and at
least one probe has succeeded since the last observed change. A
failed probe MUST NOT reset or advance the idle clock. Activity
in any descendant session MUST count as progress for the child,
including growth of a running tool's streamed `state.metadata`
(for example bash output). The probe MUST walk the subtree
depth-first, newest child first, so an active grandchild is
reached even when the child has more finished direct children
than the session cap. A probe that completes after the child's
prompt returned MUST NOT abort the child or change the run's
result.

Progress accounting MUST exclude compaction events:
context-compression summaries and synthetic continuation prompts
injected into the subtree MUST NOT count as activity changes for
the idle clock. The probe state MUST account compaction events
per session -- each session's compactions accrue since that
session's last substantive progress -- and MUST expose, for the
sessions the compaction-storm detector counts, the highest
per-session run of consecutive compactions together with the
identity of the session holding it, so the idle watchdog and the
compaction-storm detector (DELEG-FR-002) share one accounting.
Compaction events observed in different sessions MUST NOT be
summed into a single count. Substantive progress is the
successful completion of a non-read tool call, a file write, or
a checkpoint record (a non-exhaustive example list); read-only
tool calls (file reads, searches, and other inspection) MUST NOT
count, and a non-read tool call counts only when it completes
successfully.

`uf_delegate_command` MUST accept `idle_timeout` (60,000 -
1,800,000 ms) and MUST default it to 600,000 ms.

#### Scenario: Active child survives the idle window

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a grandchild session that posts new parts every
  minute
- **WHEN** 15 minutes pass
- **THEN** the delegation is not aborted

#### Scenario: Stalled child aborted early

- **GIVEN** a delegation with a 4-hour wall clock and a
  10-minute idle timeout
- **WHEN** no session in the subtree changes for 10 minutes
- **THEN** the child is aborted with `idle_timeout`

#### Scenario: Long command still printing

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a child running a 15-minute bash command whose
  `state.metadata.output` grows every minute
- **WHEN** 15 minutes pass
- **THEN** the delegation is not aborted

#### Scenario: Probe unavailable

- **GIVEN** a host whose activity probe always fails
- **WHEN** the idle window elapses
- **THEN** the child is not aborted by the idle clock

#### Scenario: Compaction-only child is stalled

- **GIVEN** a delegation with a 10-minute idle timeout
- **AND** a child whose only subtree changes during the window
  are compaction summaries
- **WHEN** the idle window elapses
- **THEN** the child is aborted with `idle_timeout` because
  compaction events do not count as progress

Previously: DT-FR-002's shared-accounting clause required the
probe state to expose a single count of compaction events
observed since the last substantive progress. Because that count
summed markers across every session of the probed subtree,
parallel read-heavy children each compacting fewer than the
storm threshold tripped the detector together (issue #1125).
This delta keeps the idle clock and every idle scenario
unchanged and refines only the shared accounting to per-session
accrual with the highest per-session run exposed alongside the
identity of the session holding it.

## REMOVED Requirements

None.
