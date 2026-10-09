## ADDED Requirements

### Requirement: Originating Issue Metadata Field

An OpenSpec change MAY include an `originating_issue`
field in its `.openspec.yaml` metadata. The field
SHALL hold a single positive integer representing a
GitHub issue number.

#### Scenario: OpenSpec change with originating issue

- **GIVEN** a developer creates a new OpenSpec change
  with `openspec new change "my-change" --issue 554`
- **WHEN** the change directory is scaffolded
- **THEN** `.openspec.yaml` SHALL contain the field
  `originating_issue: 554`

#### Scenario: OpenSpec change without originating issue

- **GIVEN** a developer creates a new OpenSpec change
  with `openspec new change "my-change"` (no
  `--issue` flag)
- **WHEN** the change directory is scaffolded
- **THEN** `.openspec.yaml` SHALL NOT contain an
  `originating_issue` field

#### Scenario: Backward compatibility with existing changes

- **GIVEN** an existing OpenSpec change whose
  `.openspec.yaml` was created before this feature
- **WHEN** `/uf.finale` reads the change metadata
- **THEN** the absence of `originating_issue` SHALL
  be treated as "no originating issue" and no
  `Closes` trailer SHALL be emitted

### Requirement: Speckit Spec Originating Issue Frontmatter

A Speckit spec MAY include an `originating_issue`
field in its `proposal.md` frontmatter. The field
SHALL hold a single positive integer representing a
GitHub issue number.

#### Scenario: Speckit spec with originating issue

- **GIVEN** a developer creates a new Speckit spec
  and includes `originating_issue: 554` in the
  `proposal.md` frontmatter
- **WHEN** `/uf.finale` reads the spec metadata
- **THEN** `/uf.finale` SHALL detect the field and
  use its value for PR body generation

#### Scenario: Speckit spec without originating issue

- **GIVEN** a Speckit spec whose `proposal.md`
  frontmatter does not contain `originating_issue`
- **WHEN** `/uf.finale` reads the spec metadata
- **THEN** the absence SHALL be treated as "no
  originating issue" and no `Closes` trailer SHALL
  be emitted

### Requirement: PR Body Closes Trailer Emission

When `/uf.finale` generates a PR body and the
originating issue number is present in the change
metadata, it SHALL append a `Closes #<N>` line to
the PR body.

#### Scenario: PR body with originating issue (OpenSpec)

- **GIVEN** the current branch is `opsx/my-change`
- **AND** `openspec/changes/my-change/.openspec.yaml`
  contains `originating_issue: 554`
- **WHEN** `/uf.finale` generates the PR body
- **THEN** the PR body SHALL contain the line
  `Closes #554` immediately after the `## Summary`
  section

#### Scenario: PR body with originating issue (Speckit)

- **GIVEN** the current branch is `NNN-my-feature`
- **AND** the Speckit spec's `proposal.md` frontmatter
  contains `originating_issue: 123`
- **WHEN** `/uf.finale` generates the PR body
- **THEN** the PR body SHALL contain the line
  `Closes #123` immediately after the `## Summary`
  section

#### Scenario: PR body without originating issue

- **GIVEN** the current branch is `opsx/my-change`
- **AND** `openspec/changes/my-change/.openspec.yaml`
  does not contain `originating_issue`
- **WHEN** `/uf.finale` generates the PR body
- **THEN** the PR body SHALL NOT contain any
  `Closes #N` or `Fixes #N` line

### Requirement: CLI Flag for Issue Number

The `openspec new change` command SHALL accept an
optional `--issue <N>` flag where `<N>` is a positive
integer.

#### Scenario: Valid issue number

- **GIVEN** the developer runs
  `openspec new change "my-change" --issue 554`
- **WHEN** the command executes successfully
- **THEN** the scaffolded `.openspec.yaml` SHALL
  contain `originating_issue: 554`

#### Scenario: Invalid issue number

- **GIVEN** the developer runs
  `openspec new change "my-change" --issue abc`
- **WHEN** the command validates its inputs
- **THEN** the command SHALL reject the input with
  a non-zero exit code and an error message
  indicating that `--issue` must be a positive
  integer

#### Scenario: Issue flag omitted

- **GIVEN** the developer runs
  `openspec new change "my-change"`
- **WHEN** the command executes successfully
- **THEN** the scaffolded `.openspec.yaml` SHALL
  NOT contain an `originating_issue` field

## MODIFIED Requirements

### Requirement: /uf.finale PR Body Generation

`/uf.finale` Step 5d (PR body generation) SHALL
read the `originating_issue` field from the change
metadata (OpenSpec `.openspec.yaml` or Speckit
`proposal.md` frontmatter) and, when present, append
a `Closes #<N>` line immediately after the
`## Summary` section of the PR body.

Previously: `/uf.finale` Step 5d generated the PR
body with Summary, How to Test, How to Demo, and Key
Files Changed sections without any originating-issue
awareness.
