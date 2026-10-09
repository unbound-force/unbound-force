## ADDED Requirements

### Requirement: Confirmation-Gate Visibility Directive

The agent MUST print the full proposed text (commit message, PR
title/body, or review verdict with inline comments) as plain
assistant output BEFORE invoking the `question` tool for any
mandatory human-confirmation gate that presents multi-line
content for approval.

The printed text MUST be identical to the content presented to
the `question` tool. The agent MUST NOT invoke the `question`
tool until the full text has been emitted as output.

#### Scenario: Commit message confirmation gate

- **GIVEN** the agent is executing `/uf.finale` Step 3
  (Generate and Confirm Commit Message)
- **WHEN** the agent has generated a proposed commit message
- **THEN** the agent MUST print the full commit message
  (summary, body, and attribution footer) as plain assistant
  output
- **AND** the agent MUST invoke the `question` tool only after
  the full commit message has been printed

#### Scenario: PR title/body confirmation gate

- **GIVEN** the agent is executing `/uf.finale` Step 5f
  (Create or Find PR — human confirmation)
- **WHEN** the agent has generated a proposed PR title and body
- **THEN** the agent MUST print the full PR title and body as
  plain assistant output
- **AND** the agent MUST invoke the `question` tool only after
  the full PR title and body have been printed

#### Scenario: Review verdict confirmation gate

- **GIVEN** the agent is executing `/uf.review-council` Step 7f
  (Verdict Mapping and Human Confirmation)
- **WHEN** the agent has assembled the review verdict and inline
  comments
- **THEN** the agent MUST print the full verdict context
  (verdict type, review body, and all inline comments) as plain
  assistant output
- **AND** the agent MUST invoke the `question` tool only after
  the full verdict context has been printed

#### Scenario: Gate content visible after context compression

- **GIVEN** a session has undergone context compression
- **WHEN** the agent re-presents a confirmation gate per the
  session-resume guard
- **THEN** the agent MUST print the full proposed text as plain
  output before invoking the `question` tool
- **AND** the printed text MUST appear in the transcript
  visible to the user

### Requirement: Post-CI Momentum Checkpoint

After Step 6 (Watch CI Checks) completes — whether checks pass,
fail, or are skipped — the agent MUST immediately mark Step 6
complete in the execution checklist and proceed to Step 7
(Return to Main) WITHOUT producing any user-facing output about
CI results.

The agent MUST defer all CI result reporting until Step 8
(Summary) is reached, where CI status is included in the
completion report.

#### Scenario: CI checks pass — continue without reporting

- **GIVEN** the agent is executing `/uf.finale` Step 6
- **WHEN** `gh pr checks` reports all checks passed
- **THEN** the agent MUST mark Step 6 complete in the execution
  checklist
- **AND** the agent MUST proceed immediately to Step 7 without
  producing user-facing output about CI results
- **AND** the agent MUST include CI pass status in the Step 8
  summary

#### Scenario: CI checks fail — continue to summary

- **GIVEN** the agent is executing `/uf.finale` Step 6
- **WHEN** `gh pr checks` reports check failures and the user
  has selected an option from the failure gate
- **THEN** the agent MUST mark Step 6 complete in the execution
  checklist
- **AND** the agent MUST proceed immediately to Step 7 without
  producing additional user-facing output about CI results
- **AND** the agent MUST include CI failure status in the Step 8
  summary

#### Scenario: Resume after compression at Step 6 boundary

- **GIVEN** a session was compressed during or immediately after
  Step 6
- **WHEN** the agent resumes and reads the execution checklist
  showing Step 6 marked complete but Steps 7-8 incomplete
- **THEN** the agent MUST proceed to Step 7 without re-reporting
  CI results
- **AND** the agent MUST complete Steps 7 and 8 before producing
  any user-facing output

### Requirement: Confirmation-Gate Visibility Audit Scope

The confirmation-gate visibility directive MUST be applied to
all mandatory human-confirmation gates that present multi-line
content for approval. Gates that present only short questions
with well-defined options (e.g., duplicate-review detection,
push confirmation) are exempt.

#### Scenario: Audit identifies Step 7f in /uf.review-council

- **GIVEN** the confirmation-gate visibility audit is performed
- **WHEN** `/uf.review-council` Step 7f presents a review
  verdict with inline comments for approval
- **THEN** Step 7f MUST include the confirmation-gate visibility
  directive

#### Scenario: Short-question gates are exempt

- **GIVEN** a `question` tool invocation presents only a short
  question string with well-defined options (no multi-line
  content)
- **WHEN** the confirmation-gate visibility audit is performed
- **THEN** the gate is exempt from the visibility directive
