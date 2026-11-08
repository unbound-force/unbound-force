## ADDED Requirements

### Requirement: Severity Scope for Code Changes

`/uf.address-feedback` MUST change code only for CRITICAL and HIGH
findings. A finding with no severity label MUST count as HIGH. A
MEDIUM or LOW finding MUST be filed as an issue, answered with a
reply, or answered with a question, and MUST NOT produce a code
change in the same run.

#### Scenario: LOW finding

- **GIVEN** an unresolved review thread whose finding is labelled LOW
- **WHEN** `/uf.address-feedback` triages it
- **THEN** the options MUST be `File issue`, `Reply`, and `Ask`
- **AND** no commit MUST address it

#### Scenario: Unlabelled human comment

- **GIVEN** a maintainer comment with no severity label
- **WHEN** it is triaged
- **THEN** it MUST follow the authority matrix as a HIGH item

#### Scenario: Filed finding

- **GIVEN** the author chooses `File issue`
- **WHEN** replies are posted
- **THEN** an issue MUST be created before the reply
- **AND** the reply MUST link the issue
- **AND** the thread MAY be resolved with the accepted threads

### Requirement: One Council per Cycle

`/uf.address-feedback` MUST NOT run `/uf.review-council` before
pushing. It MUST run the project's build, test, and lint commands
and MUST NOT push while they fail.

#### Scenario: Verification fails three times

- **GIVEN** committed fixes whose tests fail
- **WHEN** three fix attempts do not make them pass
- **THEN** the command MUST stop without pushing
