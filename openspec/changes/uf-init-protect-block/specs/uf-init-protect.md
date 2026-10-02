## ADDED Requirements

### Requirement: UF-INIT-PROTECT-001 — `/uf.init` instruction body MUST be wrapped in a `<protect>` block

The `/uf.init` command file (`.opencode/commands/uf.init.md`) MUST wrap
its core instruction body in a single `<protect>` block. The block MUST
begin immediately after the `## Instructions` heading and MUST extend
through the end of Step 12 (Legacy Directory Cleanup), before the
`### Next Steps` section.

#### Scenario: Initial protection applied
- **GIVEN** the file `.opencode/commands/uf.init.md` exists
- **AND** the file does NOT contain a `<protect>` tag between the
  `## Instructions` heading and the `### Step 0` heading
- **WHEN** `/uf.init` is executed
- **THEN** the file MUST contain a `<protect>` opening tag immediately
  after the `## Instructions` heading
- **AND** the file MUST contain a `</protect>` closing tag after the
  Step 12 content and before the `### Next Steps` heading

#### Scenario: Protection already present (idempotency)
- **GIVEN** the file `.opencode/commands/uf.init.md` exists
- **AND** the file already contains a `<protect>` tag between the
  `## Instructions` heading and the `### Step 0` heading
- **WHEN** `/uf.init` is executed
- **THEN** the file MUST NOT contain a second `<protect>` tag
- **AND** the command MUST report `⊘ uf.init.md: protect block already present (skipped)`

### Requirement: UF-INIT-PROTECT-002 — Files created by `/uf.init` Step 5 MUST include a `<protect>` block

Files created by Step 5 of `/uf.init` (speckit custom commands:
`speckit.analyze.md`, `speckit.checklist.md`, `speckit.clarify.md`,
`speckit.taskstoissues.md`) MUST include a `<protect>` block wrapping
their core instruction body at creation time. The `<protect>` block
MUST begin after the command's frontmatter and title heading, and MUST
extend through the end of the command's workflow instructions, before
the `## Guardrails` section.

#### Scenario: New speckit command created with protection
- **GIVEN** the file `.opencode/commands/speckit.analyze.md` does NOT exist
- **WHEN** `/uf.init` is executed and reaches Step 5
- **THEN** the created file MUST contain a `<protect>` block wrapping
  the core instruction body

#### Scenario: Speckit command already exists (skipped)
- **GIVEN** the file `.opencode/commands/speckit.analyze.md` already exists
- **WHEN** `/uf.init` is executed and reaches Step 5
- **THEN** the file MUST NOT be modified
- **AND** the command MUST report `⊘ speckit.analyze.md: already exists (skipped)`

### Requirement: UF-INIT-PROTECT-003 — Insertions into files with existing `<protect>` blocks MUST be placed inside the block

When `/uf.init` inserts content into a target file that already contains
a `<protect>` block (Steps 2-4, 6, 8, 10), the inserted content MUST
be placed inside the existing `<protect>` block. The insertion point
MUST be after the `<protect>` opening tag and before the `</protect>`
closing tag.

#### Scenario: Insertion into a protected file
- **GIVEN** a target file contains a `<protect>` block
- **AND** `/uf.init` needs to insert content (e.g., branch enforcement,
  guardrails, STOP HERE block) into that file
- **WHEN** `/uf.init` performs the insertion
- **THEN** the inserted content MUST be placed between the `<protect>`
  and `</protect>` tags
- **AND** the existing idempotency checks for the insertion MUST still
  apply

#### Scenario: Insertion into an unprotected file
- **GIVEN** a target file does NOT contain a `<protect>` block
- **AND** `/uf.init` needs to insert content into that file
- **WHEN** `/uf.init` performs the insertion
- **THEN** the inserted content MUST be placed at the location specified
  by the existing step instructions (unchanged behavior)
- **AND** `/uf.init` MUST NOT add a `<protect>` block to the target file

## MODIFIED Requirements

### Requirement: UF-INIT-STEP5-CREATE — Step 5 file creation templates include `<protect>` blocks

The insertion templates for Step 5 (speckit custom commands) MUST include
a `<protect>` block wrapping the core instruction body of each created
file. Previously: the templates did not include `<protect>` blocks.

#### Scenario: speckit.analyze.md created with protection
- **GIVEN** `/uf.init` creates `.opencode/commands/speckit.analyze.md`
- **WHEN** the file is written
- **THEN** the file content MUST include a `<protect>` block after the
  title heading and before the `## Guardrails` section

#### Scenario: speckit.checklist.md created with protection
- **GIVEN** `/uf.init` creates `.opencode/commands/speckit.checklist.md`
- **WHEN** the file is written
- **THEN** the file content MUST include a `<protect>` block after the
  title heading and before the `## Guardrails` section

#### Scenario: speckit.clarify.md created with protection
- **GIVEN** `/uf.init` creates `.opencode/commands/speckit.clarify.md`
- **WHEN** the file is written
- **THEN** the file content MUST include a `<protect>` block after the
  title heading and before the `## Guardrails` section

#### Scenario: speckit.taskstoissues.md created with protection
- **GIVEN** `/uf.init` creates `.opencode/commands/speckit.taskstoissues.md`
- **WHEN** the file is written
- **THEN** the file content MUST include a `<protect>` block after the
  title heading and before the `## Guardrails` section

## REMOVED Requirements

None.
