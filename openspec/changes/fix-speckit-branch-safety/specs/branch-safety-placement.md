## ADDED Requirements

### Requirement: Pre-conditions section ordering

The `speckit-workflow/SKILL.md` file MUST contain a
"Pre-conditions" section that appears before "Reading tasks.md".

#### Scenario: Agent loads speckit-workflow skill
- **GIVEN** an agent loads the `speckit-workflow` skill
- **WHEN** the agent processes the file sequentially
- **THEN** the agent MUST encounter branch safety rules before
  any workflow steps (Reading tasks.md, Creating Swarm Work
  Items, Worker Instructions, Phase Checkpoints)

#### Scenario: Branch safety rules as pre-conditions
- **GIVEN** the "Pre-conditions" section exists in the skill file
- **WHEN** the agent reads the pre-conditions
- **THEN** the section MUST contain the CRITICAL rule that all
  work MUST be committed and pushed before any branch switch
- **AND** the section MUST state that agents MUST check
  `git status --short` for uncommitted changes before creating
  a new feature branch
- **AND** the section MUST state that agents MUST never silently
  switch branches with a dirty working tree

## MODIFIED Requirements

### Requirement: Branch safety section location

The branch safety rules MUST appear in the "Pre-conditions"
section of `speckit-workflow/SKILL.md`, positioned before
"Reading tasks.md".

Previously: Branch safety rules appeared in a standalone
"Branch Safety" section at the end of the file (after all
workflow steps).

#### Scenario: No duplicate branch safety content
- **GIVEN** the branch safety rules have been moved to
  Pre-conditions
- **WHEN** the full file is inspected
- **THEN** there MUST NOT be a second "Branch Safety" section
  at the end of the file
- **AND** the branch safety content MUST appear exactly once

## REMOVED Requirements

### Requirement: Standalone "Branch Safety" section at end of file

The standalone "Branch Safety" section (previously lines 114-129)
is removed. Its content is relocated to the "Pre-conditions"
section. Removal prevents duplication and ensures agents do not
encounter the rules too late in the workflow.
