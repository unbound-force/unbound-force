## ADDED Requirements

### Requirement: Issue Triage Command

FR-001: The `/triage-issue` command MUST accept a single
GitHub issue number as a required argument. The argument
MUST be validated as a positive integer (matching
`^[1-9][0-9]*$`) before any `gh` CLI invocation.
Non-integer, negative, zero, or excessively large values
MUST be rejected with an error message.

FR-002: The command MUST validate that the issue exists
and is in an open state before proceeding. If the issue
is closed, the command MUST stop with an error message.

FR-003: The command MUST detect the current repository
dynamically via `gh repo view --json nameWithOwner`.
Repository identifiers MUST NOT be hardcoded.

#### Scenario: Triage a valid open issue

- **GIVEN** the user invokes `/triage-issue 42`
- **AND** issue #42 exists and is open
- **WHEN** the command completes all four phases
- **THEN** the issue has an appropriate label applied
- **AND** a triage comment is posted (after user
  confirmation)
- **AND** a JSON artifact is produced at
  `.uf/artifacts/issue-triage/issue-42.json`
- **AND** the artifact validates against
  `schemas/issue-triage/v1.0.0.schema.json`

#### Scenario: Triage a closed issue

- **GIVEN** the user invokes `/triage-issue 99`
- **AND** issue #99 is closed
- **WHEN** the command validates the issue state
- **THEN** the command stops with an error:
  "Issue #99 is closed. Triage applies to open issues
  only."

#### Scenario: Triage a nonexistent issue

- **GIVEN** the user invokes `/triage-issue 9999`
- **AND** issue #9999 does not exist
- **WHEN** the command fetches the issue
- **THEN** the command stops with the `gh` error output

#### Scenario: Invalid issue number argument

- **GIVEN** the user invokes `/triage-issue "42; echo pwned"`
- **WHEN** the command validates the argument
- **THEN** the command stops with an error:
  "Invalid issue number. Must be a positive integer."
- **AND** no `gh` CLI commands are executed

---

### Requirement: Multi-Agent Assessment

FR-004: The command MUST fan out to the following five
Divisor agents in parallel via the Task tool:
`divisor-adversary`, `divisor-architect`,
`divisor-guard`, `divisor-sre`, `divisor-testing`.

FR-005: Each agent MUST return a structured assessment
containing: verdict (VALID, INVALID, or
NEEDS-CLARIFICATION), category, objectivity
classification, evidence-based reasoning, and an
optional split recommendation.

FR-006: Agent discovery MUST be dynamic (read
`.opencode/agents/` directory). If an agent file is
missing, the command MUST proceed with available agents
and note which agents were unavailable.

FR-007: The command SHOULD proceed with as few as one
available agent. Zero available agents MUST cause the
command to stop with an error.

#### Scenario: All five agents available

- **GIVEN** all five Divisor agent files exist
- **WHEN** the command fans out to agents
- **THEN** all five agents are invoked in parallel
- **AND** five assessments are collected

#### Scenario: Graceful degradation with missing agents

- **GIVEN** only `divisor-adversary` and
  `divisor-architect` agent files exist
- **WHEN** the command discovers available agents
- **THEN** the command proceeds with two agents
- **AND** the artifact `summary.agents_available` is 2
- **AND** the artifact `summary.agents_consulted` is 2
- **AND** missing agents are noted in the output

#### Scenario: No agents available

- **GIVEN** no `divisor-*.md` files exist in
  `.opencode/agents/`
- **WHEN** the command attempts agent discovery
- **THEN** the command stops with an error:
  "No Divisor agents found. At least one agent is
  required for triage."

---

### Requirement: Verdict Consolidation

FR-008: The command MUST use majority consensus (3 of 5
agents) to determine issue validity. When fewer than 5
agents are available, majority of available agents
applies. Verdict resolution follows three rules in
order:
1. If NEEDS-CLARIFICATION verdicts constitute a majority
   of all agents (>50%), the overall verdict is
   NEEDS-CLARIFICATION.
2. Otherwise, NEEDS-CLARIFICATION verdicts are excluded.
   If VALID or INVALID has a majority of the remaining
   votes, that verdict wins.
3. If the remaining votes tie (equal VALID and INVALID
   after excluding NEEDS-CLARIFICATION), the overall
   verdict defaults to NEEDS-CLARIFICATION.

FR-009: Category resolution MUST use a specificity
hierarchy for non-meta categories:
bug > feature > enhancement > needs-clarification >
opinion > question. When agents disagree, the most
specific category wins. `duplicate` is resolved by
FR-013 independently of this hierarchy and takes
precedence when its conditions are met.

FR-010: Objectivity MUST be classified as `objective`
if ANY agent provides verifiable evidence. Objectivity
MUST be classified as `subjective` only when ALL agents
agree the issue is preference-based.

FR-011: Dissenting agent verdicts MUST be recorded in
the artifact with the agent name and reasoning.

#### Scenario: Majority verdict determines validity

- **GIVEN** three agents return VALID and two return
  INVALID
- **WHEN** verdicts are consolidated
- **THEN** the issue validity is VALID
- **AND** the two dissenting agents are recorded in
  `summary.dissenting_agents`

#### Scenario: Category specificity resolution

- **GIVEN** two agents classify as `bug`, two as
  `enhancement`, one as `question`
- **WHEN** categories are consolidated
- **THEN** the final category is `bug` (most specific)

#### Scenario: NEEDS-CLARIFICATION verdict

- **GIVEN** three agents return NEEDS-CLARIFICATION
  and two return VALID
- **WHEN** verdicts are consolidated
- **THEN** the issue validity is NEEDS-CLARIFICATION
  (NEEDS-CLARIFICATION has majority of all agents)
- **AND** the comment uses the needs-clarification tone
- **AND** the `needs-info` label is applied

#### Scenario: Tie-breaking with even agent count

- **GIVEN** four agents are available
- **AND** two return VALID and two return INVALID
- **WHEN** verdicts are consolidated
- **THEN** the issue validity defaults to
  NEEDS-CLARIFICATION

---

### Requirement: Duplicate Detection

FR-012: The command MUST search for potential duplicates
during the Ingest phase using
`gh issue list --search "<keywords>" --state open`.
Keywords extracted from issue titles and bodies MUST be
sanitized before use: shell metacharacters (`;`, `|`,
`` ` ``, `$()`, `"`) MUST be removed, and strings
starting with `--` MUST be stripped to prevent CLI flag
injection. Keyword extraction SHOULD prioritize the
issue title and first paragraph of the body, limited
to 5-10 keywords.

FR-013: An issue MUST be classified as `duplicate` only
when Phase 1 search finds matching candidates AND at
least two agents independently classify it as duplicate.

FR-014: Before creating child issues (during splitting),
the command MUST search for existing issues that match
each proposed child issue title.

#### Scenario: Duplicate detected

- **GIVEN** the issue title contains "timeout on large
  files"
- **AND** an open issue #30 titled "Upload timeout for
  large files" exists
- **AND** three agents classify the issue as duplicate
- **WHEN** verdicts are consolidated
- **THEN** the category is `duplicate`
- **AND** `duplicate_of` references issue #30

#### Scenario: Similar but not duplicate

- **GIVEN** keyword search returns candidate issues
- **BUT** only one agent classifies as duplicate
- **WHEN** verdicts are consolidated
- **THEN** the category is NOT `duplicate`
- **AND** the similar issues are mentioned in the
  comment for the reporter's awareness

---

### Requirement: Label Application

FR-015: The command MUST apply GitHub labels
automatically without user confirmation, using this
mapping. Exception: the `duplicate` label MUST require
user confirmation before application because it carries
implicit "close" semantics.

| Category              | Label              |
|-----------------------|--------------------|
| `bug`                 | `bug`              |
| `feature`             | `enhancement`      |
| `enhancement`         | `enhancement`      |
| `question`            | `question`         |
| `opinion`             | `design-discussion`|
| `duplicate`           | `duplicate`        |
| `needs-clarification` | `needs-info`       |

FR-016: If a required label does not exist in the
repository, the command MUST create it via
`gh label create` before applying it. If label creation
fails due to insufficient permissions, the command MUST
report the specific label that could not be created,
skip that label, and continue with remaining triage
actions. The artifact MUST record the failure in
`actions_taken`.

#### Scenario: Label applied automatically

- **GIVEN** the consolidated category is `bug`
- **WHEN** the Act phase applies labels
- **THEN** the `bug` label is applied to the issue
  via `gh issue edit <N> --add-label bug`
- **AND** no user confirmation is required

#### Scenario: Missing label created

- **GIVEN** the `design-discussion` label does not
  exist in the repository
- **AND** the consolidated category is `opinion`
- **WHEN** the Act phase applies labels
- **THEN** the label is created via `gh label create`
- **AND** the label is then applied to the issue

---

### Requirement: Comment Posting

FR-017: The command MUST compose a triage comment and
present it to the user for confirmation before posting.

FR-018: Comment text MUST be written to a temporary
file and posted via `gh api --input <file>`. Comment
text MUST NOT be interpolated into shell arguments.
Temporary files MUST be cleaned up in all exit paths
(success, user abort, API failure). Temp files MUST
use the system temp directory (`mktemp`) with
restrictive permissions (0o600).

FR-019: Comments for invalid or opinion-classified
issues MUST use a warm, non-dismissive tone that:
acknowledges the reporter's effort, explains the
reasoning with specific references, offers alternatives
when possible, and invites continued engagement.

FR-020: Comments for valid issues MUST be factual and
concise, presenting the classification and any
recommendations.

FR-021: Comments MUST include a footer:
`_This triage was performed by the Divisor review
panel._`

#### Scenario: User confirms comment

- **GIVEN** the command has composed a triage comment
- **WHEN** the user approves the comment
- **THEN** the comment is posted to the issue
- **AND** the temp file is deleted

#### Scenario: User modifies comment

- **GIVEN** the command has composed a triage comment
- **WHEN** the user chooses MODIFY
- **THEN** the user provides an adjusted comment
- **AND** the adjusted comment is posted

#### Scenario: User aborts comment

- **GIVEN** the command has composed a triage comment
- **WHEN** the user chooses ABORT
- **THEN** no comment is posted
- **AND** the artifact records
  `actions_taken.comment_posted` as false

---

### Requirement: Issue Splitting

FR-022: If two or more agents recommend splitting the
issue, the command MUST synthesize their recommendations
into proposed child issues.

FR-023: Each proposed child issue MUST be presented to
the user for confirmation before creation.

FR-024: Each child issue body MUST include a
cross-reference to the parent issue (`Split from #N`).
Child issue titles and bodies MUST be written to
temporary files and created via `gh api --input <file>`
or equivalent. Child issue content MUST NOT be
interpolated into shell arguments.

FR-025: After creating child issues, the command MUST
post a comment on the parent issue listing all created
child issues with their numbers and titles.

FR-026: The parent issue MUST NOT be auto-closed after
splitting.

#### Scenario: Issue split into two child issues

- **GIVEN** three agents recommend splitting issue #42
  into "Fix timeout handling" and "Add retry
  configuration"
- **AND** the user confirms both child issues
- **WHEN** the Act phase creates child issues
- **THEN** two new issues are created with
  cross-references to #42
- **AND** a comment is posted on #42 listing the new
  issues
- **AND** issue #42 remains open

---

### Requirement: Triage Artifact

FR-027: The command MUST produce a JSON artifact at
`.uf/artifacts/issue-triage/issue-<N>.json` for every
invocation, regardless of outcome.

FR-028: The artifact MUST be wrapped in the standard
envelope format with `artifact_type: "issue-triage"`.

FR-029: The artifact payload MUST conform to the
`schemas/issue-triage/v1.0.0.schema.json` schema.

FR-030: The artifact MUST include per-agent assessments
with full reasoning, enabling provenance tracing of
every classification decision.

#### Scenario: Artifact produced on successful triage

- **GIVEN** the command completes all four phases
- **WHEN** the artifact is written
- **THEN** the file exists at
  `.uf/artifacts/issue-triage/issue-<N>.json`
- **AND** the payload validates against the schema
- **AND** the envelope contains `hero`, `version`,
  `timestamp`, `artifact_type`, `schema_version`,
  and `context`

#### Scenario: Artifact produced on abort

- **GIVEN** the user aborts during the Act phase
- **WHEN** the artifact is written
- **THEN** `actions_taken.comment_posted` is false
- **AND** `actions_taken.child_issues_created` is empty
- **AND** the assessments are still recorded

---

### Requirement: Guardrails

FR-031: The command MUST NOT close or lock any issue
under any circumstances.

FR-032: The command MUST NOT post comments without
explicit user confirmation.

FR-033: The command MUST NOT create child issues without
explicit user confirmation.

FR-034: The command MUST process exactly one issue per
invocation.

FR-035: The command MUST verify `gh` CLI installation
(via `which gh`) and authentication (via
`gh auth status`) before any GitHub API calls. If `gh`
is not installed: "GitHub CLI (gh) is not installed.
Install from https://cli.github.com/". If not
authenticated: "GitHub CLI not authenticated. Run
`gh auth login` to authenticate."

FR-036: When any `gh` CLI or API call fails during any
phase (network error, HTTP 403 rate limit, HTTP 5xx),
the command MUST report the specific error, indicate
which phase failed, and list any actions already
completed. The command MUST NOT proceed with subsequent
GitHub mutations after a failure but MUST still produce
the artifact with `actions_taken` reflecting partial
state.

FR-037: The command MUST be safely re-runnable. On
re-invocation for the same issue, the command MUST
detect previously applied labels (via `gh issue view
--json labels`) to avoid duplication. Previously posted
triage comments SHOULD be detected by checking for the
Divisor review panel footer. If an artifact already
exists for the issue, the command MUST append a round
number (e.g., `issue-42-2.json`) to preserve history.

FR-038: All untrusted text (issue content, agent
output, synthesized comments, child issue content)
MUST be written to temporary files and passed via
`--input` for all `gh api` and `gh issue create` calls.
Untrusted text MUST NOT be interpolated into shell
arguments. Temp files MUST use restrictive permissions
(0o600) and be cleaned up in all exit paths.

FR-039: Artifact file paths MUST be constructed safely.
The issue number MUST be validated as a positive integer
(FR-001) before use in any file path.

#### Scenario: No auto-close

- **GIVEN** the consolidated verdict is INVALID
- **WHEN** the Act phase completes
- **THEN** the issue remains open
- **AND** only a label and comment (if confirmed) are
  applied

## MODIFIED Requirements

None.

## REMOVED Requirements

None.
