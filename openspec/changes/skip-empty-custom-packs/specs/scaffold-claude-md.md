## ADDED Requirements

### Requirement: FR-01 Empty Custom Pack Detection

The scaffold engine MUST provide a helper function that
determines whether a convention pack file on disk contains
actual rule content. A file is considered to have rule content
if and only if there is at least one non-whitespace character
after the last occurrence of the placeholder sentinel string
`<!-- Add project-specific rules below this line -->`. If the
file does not exist or cannot be read, the helper MUST return
`true` (fail-open).

#### Scenario: Empty stub file returns false

- **GIVEN** a custom pack file that contains only the
  boilerplate scaffold (YAML frontmatter, section heading,
  description, `## Custom Rules` heading, and the placeholder
  HTML comment) with no content after the placeholder
- **WHEN** `hasRuleContent` is called with the file path
- **THEN** it returns `false`

#### Scenario: Populated file returns true

- **GIVEN** a custom pack file that contains one or more
  non-whitespace characters after the placeholder comment
- **WHEN** `hasRuleContent` is called with the file path
- **THEN** it returns `true`

#### Scenario: Missing file returns true (fail-open)

- **GIVEN** a file path that does not exist on disk
- **WHEN** `hasRuleContent` is called with that path
- **THEN** it returns `true`

### Requirement: FR-02 Conditional Custom Pack Import

When generating the managed `CLAUDE.md` block, the scaffold
engine MUST omit the `@.opencode/uf/packs/<name>` import line
for any `*-custom.md` pack whose corresponding file on disk
satisfies the empty detection rule (FR-01). Non-custom packs
(e.g., `default.md`, `go.md`) MUST always be included
regardless of content.

#### Scenario: All custom packs empty — imports omitted

- **GIVEN** a project root where all deployed `*-custom.md`
  files contain no rule content (per FR-01)
- **WHEN** `uf init` runs and generates the managed `CLAUDE.md`
  block
- **THEN** the generated block contains no `@` import lines
  for any `*-custom.md` pack

#### Scenario: One custom pack populated — only it is imported

- **GIVEN** a project root where `go-custom.md` contains rule
  content and `default-custom.md` and `content-custom.md` are
  empty stubs
- **WHEN** `uf init` runs and generates the managed `CLAUDE.md`
  block
- **THEN** the generated block contains an import for
  `go-custom.md` and no imports for `default-custom.md` or
  `content-custom.md`

#### Scenario: Root not provided — all packs included (no-op)

- **GIVEN** `collectDeployedPacks` is called without a project
  root (empty string)
- **WHEN** the function executes
- **THEN** all packs including custom packs are returned
  (backward-compatible behaviour)

### Requirement: FR-03 Idempotent Re-evaluation

When `uf init --reinit` is run after a user has added content
to a previously empty custom pack, the managed `CLAUDE.md`
block MUST be updated to include the now-populated custom pack
import. This MUST be automatic with no additional user action
required beyond running `uf init --reinit`.

#### Scenario: Custom pack gains content after initial init

- **GIVEN** a repo where `uf init` was previously run and
  `default-custom.md` was empty (import omitted from CLAUDE.md)
- **WHEN** the user adds rules to `default-custom.md` and
  runs `uf init --reinit`
- **THEN** the managed `CLAUDE.md` block is updated to include
  `@.opencode/uf/packs/default-custom.md`

## MODIFIED Requirements

### Requirement: CLAUDE.md Managed Block Generation

Previously: `buildCLAUDEmdBlock(lang string) string` included
all custom packs unconditionally.

The function signature SHALL be updated to
`buildCLAUDEmdBlock(lang, root string) string` where `root`
is the project root directory. The function MUST pass `root`
to `collectDeployedPacks` so that empty custom packs are
excluded when a root is provided.

Similarly, `collectDeployedPacks(lang string) []string` SHALL
be updated to `collectDeployedPacks(lang, root string)
[]string`. When `root == ""`, the function MUST return the
same list as before (all packs). When `root != ""`, the
function MUST filter out custom packs that fail FR-01.

## REMOVED Requirements

None.
