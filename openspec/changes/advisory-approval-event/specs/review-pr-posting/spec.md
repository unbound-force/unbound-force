## ADDED Requirements

### Requirement: Verdict to GitHub Event Mapping

`/uf.review-pr` MUST map `APPROVE WITH ADVISORIES` to the `APPROVE`
event when no finding is CRITICAL or HIGH, and to the `COMMENT`
event otherwise. A finding with no recognized severity MUST count
as HIGH for this mapping.

#### Scenario: Advisories are MEDIUM or LOW

- **GIVEN** a verdict of `APPROVE WITH ADVISORIES`
- **WHEN** every finding is MEDIUM or LOW
- **THEN** the review MUST post with the `APPROVE` event
- **AND** the advisories MUST remain in the review

#### Scenario: An advisory is HIGH or CRITICAL

- **GIVEN** a verdict of `APPROVE WITH ADVISORIES`
- **WHEN** any finding is CRITICAL or HIGH
- **THEN** the review MUST post with the `COMMENT` event
- **AND** the review body MUST say those findings need human
  judgment before merge

#### Scenario: A severity is missing

- **GIVEN** a verdict of `APPROVE WITH ADVISORIES`
- **WHEN** a finding has no recognized severity
- **THEN** the review MUST post with the `COMMENT` event

#### Scenario: Self-authored pull request

- **GIVEN** the viewer is the pull request author
- **WHEN** the mapped event is `APPROVE` or `REQUEST_CHANGES`
- **THEN** the review MUST post with the `COMMENT` event
