<!-- capability: delegation-timeouts
     change: compaction-storm-on-a-tight-context-mode
     originating_issue: 1079
     rationale: refines DT-FR-002 (Idle Timeout) so
     compaction events no longer count as probe progress;
     the storm detector consuming the shared accounting
     lives in the new delegation capability (DELEG-FR-002)
-->

Requirement numbering: this change modifies DT-FR-002
(Idle Timeout); no new delegation-timeouts ids are
allocated, and the DT-FR-005 through DT-FR-012 reservation
from openspec/changes/delegation-timeout-budgets/ is
unchanged. The sibling capability `delegation` (new in
this change) allocates DELEG-FR-001 through DELEG-FR-003
and reserves DELEG-FR-004 through DELEG-FR-018. Under
OpenSpec replacement semantics this MODIFIED delta
replaces the requirement's prose and scenarios; the
predecessor scenarios are carried forward and joined by
the new compaction-exclusion scenario. Requirement ids
MUST stay inside the allocated range.

## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: Idle Timeout

**ID: DT-FR-002** This requirement MUST be satisfied.

When an idle timeout is configured, the executor MUST poll
the child's session subtree for activity and MUST abort the
child with error code `idle_timeout` (retryable) only after
no activity change has been observed for the idle window
and at least one probe has succeeded since the last
observed change. A failed probe MUST NOT reset or advance
the idle clock. Activity in any descendant session MUST
count as progress for the child, including growth of a
running tool's streamed `state.metadata` (for example bash
output). The probe MUST walk the subtree depth-first,
newest child first, so an active grandchild is reached even
when the child has more finished direct children than the
session cap. A probe that completes after the child's
prompt returned MUST NOT abort the child or change the
run's result.

Progress accounting MUST exclude compaction events:
context-compression summaries and synthetic continuation
prompts injected into the subtree MUST NOT count as
activity changes for the idle clock. The probe state MUST
expose the count of compaction events observed since the
last substantive progress, so the idle watchdog and the
compaction-storm detector (DELEG-FR-002) share one
accounting. Substantive progress is the successful
completion of a non-read tool call, a file write, or a
checkpoint record (a non-exhaustive example list);
read-only tool calls (file reads, searches, and other
inspection) MUST NOT count, and a non-read tool call
counts only when it completes successfully.

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
- **WHEN** no session in the subtree changes for 10
  minutes
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
- **AND** a child whose only subtree changes during the
  window are compaction summaries
- **WHEN** the idle window elapses
- **THEN** the child is aborted with `idle_timeout`
  because compaction events do not count as progress

Previously: DT-FR-002 counted any activity change in the
child's subtree as progress for idle-timeout purposes,
including compaction summaries and synthetic continuation
prompts, so a child in a compaction loop appeared active
and was never aborted by the idle clock (issue #1079
incident: roughly 150 compactions in 110 minutes).

## REMOVED Requirements

None.
