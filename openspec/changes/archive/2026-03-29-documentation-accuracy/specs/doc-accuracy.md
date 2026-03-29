## ADDED Requirements

None.

## MODIFIED Requirements

### Requirement: readme-knowledge-layer

README.md MUST reference Dewey as the project's
knowledge layer. All graphthulhu references in README
MUST be replaced with Dewey equivalents.

Previously: README.md referenced graphthulhu as the
knowledge graph server with a link to the
skridlevsky/graphthulhu repository.

#### Scenario: README knowledge layer accuracy
- **GIVEN** a contributor reads README.md
- **WHEN** they look at the Knowledge Graph section
- **THEN** the section describes Dewey (not
  graphthulhu) as the knowledge layer, references the
  `unbound-force/dewey` repository, and describes MCP
  semantic search capabilities

### Requirement: readme-counts

README.md MUST state the correct number of
architectural specifications and scaffold files.

Previously: README stated "10 architectural
specifications" and "47 files."

#### Scenario: README counts accuracy
- **GIVEN** a contributor reads README.md
- **WHEN** they see spec and file counts
- **THEN** the spec count matches the actual number of
  spec directories under `specs/` (16) and the scaffold
  file count matches the test assertion in
  `scaffold_test.go` (50)

### Requirement: agents-hero-table-binary

The AGENTS.md hero table MUST reference the correct
binary name (`unbound-force`) for embedded heroes.

Previously: Hero table said "Embedded in `unbound`
binary" for Cobalt-Crush and The Divisor.

#### Scenario: Hero table binary name
- **GIVEN** a contributor reads the hero table in
  AGENTS.md
- **WHEN** they see the Repo column for embedded heroes
- **THEN** it says `unbound-force` binary, not
  `unbound` binary

### Requirement: agents-hero-status

The AGENTS.md hero status description MUST reflect
the current implementation state of all heroes.

Previously: Text said "Gaze is the only hero with a
functional implementation."

#### Scenario: Hero implementation status
- **GIVEN** a contributor reads the hero status in
  AGENTS.md
- **WHEN** they look below the hero table
- **THEN** the text accurately describes all five
  heroes as implemented, with Gaze noted as having
  the most mature standalone implementation

### Requirement: agents-project-structure

The AGENTS.md project structure tree MUST include all
existing spec directories.

Previously: Tree jumped from 013 to 016, omitting
014-dewey-architecture and 015-dewey-integration.

#### Scenario: Project structure completeness
- **GIVEN** a contributor reads the project structure
  in AGENTS.md
- **WHEN** they look at the specs/ section
- **THEN** all spec directories including
  014-dewey-architecture and 015-dewey-integration
  are listed

### Requirement: agents-spec-framework-binary

The AGENTS.md Specification Framework section MUST
reference the correct binary name.

Previously: Text said "distributed via the `unbound`
CLI binary."

#### Scenario: Spec framework binary reference
- **GIVEN** a contributor reads the Specification
  Framework section
- **WHEN** they see the CLI binary reference
- **THEN** it says `unbound-force` CLI binary (or `uf`)

### Requirement: agents-phase-descriptions

The AGENTS.md spec phase descriptions SHOULD reflect
current state where practical.

Previously: Phase 0 said "Three core principles"
(now four). Phase 1 said "Three-persona review
protocol" for Divisor (now five personas).

#### Scenario: Phase description accuracy
- **GIVEN** a contributor reads spec phase descriptions
- **WHEN** they see principle counts or persona counts
- **THEN** the numbers match the current state (four
  principles, five personas)

### Requirement: agents-sibling-repos

The AGENTS.md Sibling Repositories table SHOULD
include all active repositories in the unbound-force
GitHub organization.

Previously: Table listed Gaze, Website, and
homebrew-tap but not Dewey.

#### Scenario: Sibling repos completeness
- **GIVEN** a contributor reads the Sibling
  Repositories table
- **WHEN** they check for the Dewey repo
- **THEN** `unbound-force/dewey` is listed with its
  purpose and status

### Requirement: spec-frontmatter-status

Spec frontmatter `status` field MUST reflect the
actual implementation state. Specs with all tasks
completed and PRs merged MUST have `status: complete`.

Previously: Specs 012-016 had `status: draft` despite
all implementation work being done.

#### Scenario: Spec status accuracy
- **GIVEN** a contributor checks spec frontmatter
- **WHEN** they read `status` for specs 012-016
- **THEN** all five show `status: complete`

## REMOVED Requirements

None.
