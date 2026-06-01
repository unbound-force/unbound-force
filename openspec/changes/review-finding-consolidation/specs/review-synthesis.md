## ADDED Requirements

### Requirement: Compound Severity Escalation

`severity.md` MUST define a Compound Severity
Escalation section that instructs personas to assess
combined severity when multiple findings share a root
cause. The rule MUST require three conditions for
consolidation: (a) same component or pipeline stage,
(b) shared root cause, (c) combined risk greater than
any individual finding. The section MUST include
example escalation patterns and an anti-consolidation
guard for independent findings.

#### Scenario: Two related supply chain findings

- **GIVEN** a PR introduces an unverified binary
  download (MEDIUM) in a `--privileged` CI container
  (MEDIUM)
- **WHEN** the reviewer applies compound severity
  escalation
- **THEN** the two findings are consolidated into a
  single HIGH finding citing both contributing factors

#### Scenario: Two unrelated findings in same file

- **GIVEN** a PR has a missing error wrap (MEDIUM) and
  an overly broad file permission (MEDIUM) in the same
  file but with independent root causes
- **WHEN** the reviewer applies compound severity
  escalation
- **THEN** the findings remain separate at MEDIUM each

### Requirement: Severity Calibration Step

`/review-pr` MUST include a Severity Calibration step
(Step 8g) after finding consolidation (Step 8f). For
each finding, the reviewer MUST re-read the matching
`severity.md` definition and quote it as evidence for
the severity assignment. If the quoted definition maps
to a different severity level, the assignment MUST be
adjusted.

#### Scenario: Anchoring bias correction

- **GIVEN** a reviewer initially classifies a finding
  as MEDIUM based on overall PR quality impression
- **WHEN** the severity calibration step forces
  re-reading the severity definition
- **THEN** the finding is reclassified to HIGH because
  the definition for HIGH explicitly matches the
  finding's characteristics

### Requirement: CI Bot Annotation Cross-referencing

`/review-pr` MUST include a CI Bot Annotation
Cross-referencing step (Step 8e) before finding
consolidation. The step MUST cross-reference inline
comments from CI bots (Scorecard, Trivy, GHAS,
Dependabot, CodeQL) against the reviewer's own
findings. Matching bot findings MUST be cited as
corroborating evidence, not dismissed as unrelated.

#### Scenario: Scorecard flags same dependency issue

- **GIVEN** Scorecard has flagged an unpinned
  dependency in a CI workflow
- **WHEN** the Adversary persona also identifies the
  same unpinned dependency
- **THEN** the reviewer cites the Scorecard finding
  as corroborating evidence and uses it to strengthen
  the severity classification

### Requirement: Finding Consolidation Pass

`/review-pr` MUST include a Finding Consolidation
step (Step 8f) after CI bot cross-referencing. The
step MUST group findings sharing a root cause using
the compound severity escalation rule from
`severity.md`. Each consolidated finding MUST list
contributing factors with original category attribution.

`/review-council` MUST include a cross-persona finding
consolidation pass in Code Review Mode (Step 3) and
Spec Review Mode (Step 2) before the fix loop. The
pass MUST group findings from different personas
sharing a root cause and apply compound severity
escalation.

#### Scenario: Cross-persona consolidation

- **GIVEN** the Adversary flags "missing checksum" as
  MEDIUM and the SRE flags "privileged blast radius"
  as MEDIUM on the same CI pipeline
- **WHEN** `/review-council` applies cross-persona
  consolidation
- **THEN** the findings are merged into a single
  consolidated finding at HIGH with attribution to
  both personas

### Requirement: Adversarial Input Enumeration

`divisor-adversary.md` Audit Checklist and
`/review-pr` Step 8b MUST include an adversarial
input enumeration substep. For each new input,
parameter, or secret introduced by the change, the
reviewer MUST enumerate: (a) what values a caller can
pass, (b) what happens for each edge case (empty,
wrong type, wrong case, injection payload), (c)
whether validation exists and is sufficient.

#### Scenario: Case-sensitive input without validation

- **GIVEN** a PR introduces a `generate_attestations`
  string parameter that is compared case-sensitively
- **WHEN** the reviewer enumerates adversarial inputs
- **THEN** the reviewer identifies that passing
  `"True"` instead of `"true"` silently skips
  attestation generation and classifies this as HIGH

### Requirement: Dependency Necessity Check

`divisor-adversary.md` Audit Checklist item 2 MUST
ask "Is each external tool, binary, or library
justified?" before "Are dependencies pinned?" The
check MUST determine whether the project's existing
toolchain could cover the same use case.

#### Scenario: Unnecessary binary download

- **GIVEN** a PR downloads an external binary for a
  task the project's existing tools can perform
- **WHEN** the reviewer applies the dependency
  necessity check
- **THEN** the reviewer flags the dependency as
  unnecessary attack surface and recommends removal

### Requirement: Issue Suggestion Gap Detection

`/review-pr` Step 8a MUST include a substep that
scans linked issue bodies for explicit code
suggestions (fenced code blocks, inline code). For
each suggestion, the reviewer MUST check whether the
PR implemented it. Unimplemented suggestions MUST be
flagged as findings with severity based on risk.

#### Scenario: Linked issue suggests guard clause

- **GIVEN** a linked issue suggests adding
  `cancel-in-progress: ${{ !github.ref_protected }}`
- **WHEN** the PR keeps unconditional
  `cancel-in-progress: true`
- **THEN** the reviewer flags this as a HIGH finding
  ("destructive operation without guard")

## ADDED Requirements (Constitution)

### Requirement: Principle V — Security by Default

The org constitution MUST include a fifth core
principle: Security by Default. The principle MUST
require: (a) supply chain integrity — dependencies
verified by content hash, (b) input validation — all
external inputs validated before use, (c) least
privilege — components operate with minimum necessary
permissions.

#### Scenario: New hero constitution alignment

- **GIVEN** a hero repository has a constitution
  aligned to org constitution v1.1.0
- **WHEN** org constitution v1.2.0 adds Principle V
- **THEN** the hero MUST open an alignment issue
  within one release cycle to review Principle V
  compliance
