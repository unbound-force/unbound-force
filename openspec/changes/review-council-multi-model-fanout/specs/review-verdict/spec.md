## ADDED Requirements

### Requirement: Review Verdict Version 2 Migration

**ID: RV-FR-001** This requirement MUST be satisfied.

The canonical Divisor decision schema MUST add
`schemas/review-verdict/v2.0.0.schema.json`. Version 1.0.0 MUST remain
available for historical reads. New producers MUST emit schema 2.0.0.
The major version MUST signal the required consumer migration.

Version 2 MUST retain the version 1 payload fields. Its closed
`council_decision` enum MUST be `APPROVED`, `CHANGES_REQUESTED`,
`ESCALATED`, `INCONCLUSIVE`, or `UNAVAILABLE`.

Projection from `review-dispatch` MUST be deterministic:

- `APPROVE` MUST become `APPROVED`;
- `APPROVE WITH ADVISORIES` MUST become `ESCALATED`;
- `REQUEST CHANGES` MUST become `CHANGES_REQUESTED`;
- `INCONCLUSIVE` MUST remain `INCONCLUSIVE`; and
- `UNAVAILABLE` MUST remain `UNAVAILABLE`.

Mx F, Cobalt-Crush, and Muti-Mind consumers MUST migrate to 2.0.0 in
the same change. They MUST block automated progression for
`CHANGES_REQUESTED`, `ESCALATED`, `INCONCLUSIVE`, and `UNAVAILABLE`
unless their existing human gate explicitly resolves the result. They
MUST NOT interpret either no-success decision as approval.

Specs 008 and 009 MUST migrate in the same change. Spec 008
orchestration MUST block `INCONCLUSIVE` and `UNAVAILABLE` until human
resolution or a successful rerun. Spec 009 MUST register the
five-value version 2 enum, retain v1 historical reads, and preserve
major-version rejection for unmigrated consumers.

Registry documentation, samples, producer and consumer tables, and
compatibility tests MUST describe the migration. Migrated consumers
MUST read v1 historical artifacts. An unmigrated v1-only consumer MUST
reject v2 through the existing major-version compatibility rule.

#### Scenario: Availability-only review failure

- **GIVEN** a review has no success due only to provider availability
- **WHEN** canonical and dispatch artifacts are emitted
- **THEN** both record `UNAVAILABLE` without lossy projection
- **AND** migrated consumers block automated progression

#### Scenario: Inconclusive persistence failure

- **GIVEN** dispatch persistence prevents an observable valid result
- **WHEN** the canonical decision is emitted successfully
- **THEN** review-verdict 2.0.0 records `INCONCLUSIVE`
- **AND** the calculated assessment remains human-only provenance

#### Scenario: Historical version remains readable

- **GIVEN** a migrated consumer reads a v1 review-verdict artifact
- **WHEN** schema compatibility is checked
- **THEN** historical reading remains supported
- **AND** new writes still use schema 2.0.0

#### Scenario: Orchestration receives no-success decision

- **GIVEN** version 2 records `INCONCLUSIVE` or `UNAVAILABLE`
- **WHEN** Spec 008 orchestration evaluates stage progression
- **THEN** automated progression remains blocked
- **AND** the decision requires human resolution or successful rerun
