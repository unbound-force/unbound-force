<!-- capability: address-feedback-scope
     change: feedback-loop-scope
     requirement ids: AFS-FR-001 to AFS-FR-003 allocated here -->

## ADDED Requirements

### Requirement: Severity Scope for Code Changes

**ID: AFS-FR-001** This requirement MUST be satisfied.

`/uf.address-feedback` MUST change code only for CRITICAL and HIGH
findings. A finding with no severity label MUST count as HIGH. A
MEDIUM or LOW finding MUST be filed as an issue, answered with a
reply, or answered with a question, and MUST NOT produce a code
change in the same run. MEDIUM and LOW items MUST be assessed at
Tier 1.

#### Scenario: LOW finding

- **GIVEN** an unresolved review thread whose finding is labelled LOW
- **WHEN** `/uf.address-feedback` triages it
- **THEN** the options MUST be `File issue`, `Reply`, and `Ask`
- **AND** no commit MUST address it

#### Scenario: Unlabelled human comment

- **GIVEN** a maintainer comment with no severity label
- **WHEN** it is triaged
- **THEN** it MUST follow the authority matrix as a HIGH item

### Requirement: Filing and Answering Findings

**ID: AFS-FR-002** This requirement MUST be satisfied.

A FILE item's issue MUST be created before its reply, from title and
body text written to temporary files. If issue creation fails twice,
the item's reply MUST NOT be posted and its thread MUST stay open.
The threads of ACCEPT, MODIFY, FILE, and REPLY items MUST be offered
for resolution. FILE and REPLY items MUST be recorded in the
feedback-triage artifact as `reject` with a `decision_reasoning`
prefix of `filed: <issue URL>` or `reply:`.

#### Scenario: Filed finding

- **GIVEN** the author chooses `File issue`
- **WHEN** replies are posted
- **THEN** an issue MUST exist before the reply
- **AND** the reply MUST link the issue

#### Scenario: Issue creation fails

- **GIVEN** `gh issue create` fails on the first try and the retry
- **WHEN** replies are posted
- **THEN** that item's reply MUST NOT be posted
- **AND** the other items MUST still be posted

### Requirement: One Council per Cycle

**ID: AFS-FR-003** This requirement MUST be satisfied.

`/uf.address-feedback` MUST NOT run `/uf.review-council` before
pushing. It MUST run the `pre-flight` skill in `hard-gate` mode and
MUST NOT push while a check fails. It MUST stop without pushing
after 3 failed attempts.

#### Scenario: Verification keeps failing

- **GIVEN** committed fixes whose tests fail
- **WHEN** three attempts do not make the checks pass
- **THEN** the command MUST stop without pushing

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
