# Feature Specification: Multi-Platform Scaffold Deployment

**Feature Branch**: `035-platform-scaffold`
**Created**: 2026-04-29
**Status**: Draft
**Input**: User description: "Multi-platform scaffold deployment for uf init with --platform flag supporting OpenCode and Cursor targets"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cursor-Only Project Scaffolding (Priority: P1)

A developer starts a new project using Cursor as their
primary AI coding assistant. They run `uf init --platform
cursor` and receive a fully native `.cursor/` directory
with agents, commands, rules, skills, and MCP
configuration -- all in Cursor's native formats. The
project works immediately in Cursor without any
additional configuration.

**Why this priority**: This is the core value
proposition. Without correct Cursor file generation, the
feature has no purpose.

**Independent Test**: Can be fully tested by running
`uf init --platform cursor` in an empty directory and
verifying all generated files match Cursor's expected
directory structure, frontmatter schemas, and file
extensions.

**Acceptance Scenarios**:

1. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `uf init --platform cursor`,
   **Then** a `.cursor/` directory is created containing
   `agents/`, `commands/`, `rules/`, and `skills/`
   subdirectories with correctly formatted files, and no
   `.opencode/` directory is created.

2. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `uf init --platform cursor`,
   **Then** agent files in `.cursor/agents/` have YAML
   frontmatter with `name` (derived from filename) and
   `description` fields, and do not contain OpenCode-
   specific fields (`mode`, `temperature`, `tools`).

3. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `uf init --platform cursor`,
   **Then** convention packs are deployed as `.mdc` files
   in `.cursor/rules/` with `description`, `globs`, and
   `alwaysApply` frontmatter fields.

4. **Given** an empty directory with a `tsconfig.json`,
   **When** a user runs `uf init --platform cursor`,
   **Then** language-specific packs (`typescript.md`,
   `typescript-custom.md`) are deployed as `.mdc` rules
   with `globs: "**/*.{ts,tsx,js,jsx}"`, and Go-specific
   packs are not deployed.

---

### User Story 2 - Dual-Platform Project Scaffolding (Priority: P2)

A team uses both OpenCode and Cursor across different
developers. A lead runs `uf init --platform opencode
--platform cursor` and the project receives both
`.opencode/` and `.cursor/` directories, each containing
platform-native files derived from the same canonical
embedded assets. Both tools work correctly in the same
project.

**Why this priority**: Multi-platform support is the
strategic differentiator over single-tool setups. Without
this, users must choose one platform or manually maintain
two sets of configuration.

**Independent Test**: Can be tested by running
`uf init --platform opencode --platform cursor` in an
empty directory and verifying both `.opencode/` and
`.cursor/` directories are created with correct,
platform-appropriate content.

**Acceptance Scenarios**:

1. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `uf init --platform opencode
   --platform cursor`, **Then** both `.opencode/` and
   `.cursor/` directories are created, each containing
   agents, commands, and platform-appropriate rule/pack
   files.

2. **Given** a project with an existing `.opencode/`
   directory from a previous `uf init`, **When** a user
   runs `uf init --platform cursor`, **Then** a
   `.cursor/` directory is created alongside the existing
   `.opencode/` directory without modifying it.

3. **Given** a dual-platform project, **When** a user
   runs `uf init --platform opencode --platform cursor`
   again, **Then** tool-owned files in both directories
   are updated if content has changed, and user-owned
   files in both directories are preserved.

---

### User Story 3 - Convention Pack to Cursor Rule Translation (Priority: P2)

A developer using Cursor wants convention packs to
appear as native Cursor rules in the Rules UI, with
automatic attachment to matching files via glob patterns.
When they open a Go file, the Go convention pack is
automatically attached to the conversation context
without manual `@`-referencing.

**Why this priority**: Glob-based auto-attachment is
Cursor's primary rule discovery mechanism. Without it,
convention packs are invisible to Cursor's built-in
rule system.

**Independent Test**: Can be tested by inspecting the
generated `.mdc` files for correct frontmatter and
verifying that Cursor's rule matching behavior activates
them for the expected file types.

**Acceptance Scenarios**:

1. **Given** a Go project scaffolded with
   `--platform cursor`, **When** Cursor opens a `.go`
   file, **Then** the `uf-go.mdc` rule in
   `.cursor/rules/` has `globs: "**/*.go"` and
   `alwaysApply: false`, so it auto-attaches.

2. **Given** a project scaffolded with
   `--platform cursor`, **Then** language-agnostic packs
   (`default`, `severity`, `content`) are deployed as
   `.mdc` rules with `alwaysApply: true` and no globs.

3. **Given** a custom convention pack
   (`go-custom.md`) that a user has modified, **When**
   `uf init --platform cursor` is run again, **Then** the
   corresponding `.cursor/rules/uf-go-custom.mdc` file
   is NOT overwritten (user-owned).

---

### User Story 4 - MCP Configuration Translation (Priority: P3)

A developer using Cursor wants MCP servers (Dewey,
Replicator) to be available in Cursor without manual
configuration. When `uf init --platform cursor` runs,
it generates a `.cursor/mcp.json` file with servers
translated from the OpenCode MCP format.

**Why this priority**: MCP servers provide critical
tooling (semantic search, swarm orchestration) but the
configuration format differs between platforms. Without
translation, Cursor users lose access to ecosystem tools.

**Independent Test**: Can be tested by running
`uf init --platform cursor` in a project where `dewey`
and `replicator` binaries are in PATH, then inspecting
`.cursor/mcp.json` for correct format.

**Acceptance Scenarios**:

1. **Given** a project where `dewey` is in PATH, **When**
   a user runs `uf init --platform cursor`, **Then**
   `.cursor/mcp.json` is created with a `mcpServers`
   key containing a `dewey` entry with `command` (string),
   `args` (array), and no `type` field (Cursor uses
   implicit stdio).

2. **Given** an existing `opencode.json` with custom MCP
   servers using `{env:VAR}` syntax, **When**
   `uf init --platform cursor` is run, **Then**
   `.cursor/mcp.json` contains the same servers with
   environment variable syntax converted to `${VAR}`.

3. **Given** neither `dewey` nor `replicator` is in PATH,
   **When** a user runs `uf init --platform cursor`,
   **Then** no `.cursor/mcp.json` is created (no servers
   to configure).

---

### User Story 5 - DivisorOnly Mode for Cursor (Priority: P3)

A developer wants to use only the Divisor review council
in Cursor without the full Unbound Force scaffold. They
run `uf init --divisor --platform cursor` and receive
only the 6 Divisor agent files, the review-council
command, and the applicable convention packs as Cursor
rules.

**Why this priority**: DivisorOnly mode already exists
for OpenCode. Extending it to Cursor is a parity feature
that enables adoption of the review council without
full commitment.

**Independent Test**: Can be tested by running
`uf init --divisor --platform cursor` and verifying only
Divisor-related files are deployed to `.cursor/`.

**Acceptance Scenarios**:

1. **Given** an empty directory with a `go.mod` file,
   **When** a user runs `uf init --divisor --platform
   cursor`, **Then** `.cursor/agents/` contains only
   `divisor-*.md` files, `.cursor/commands/` contains
   only `review-council.md`, and `.cursor/rules/`
   contains only the applicable convention pack rules.

2. **Given** DivisorOnly mode with `--platform cursor`,
   **Then** no OpenSpec directories, skills, or non-
   Divisor agents are created.

---

### Edge Cases

- What happens when `--platform` is given an unknown
  value (e.g., `--platform vim`)? System rejects with
  a clear error listing valid platforms.
- What happens when `--platform` is specified multiple
  times with the same value (e.g., `--platform cursor
  --platform cursor`)? Deduplicated silently; files
  deployed once.
- What happens when a `.cursor/agents/` file exists
  with content the user has customized? User-owned
  files are never overwritten (same ownership model as
  `.opencode/`).
- What happens when `opencode.json` has MCP servers
  with remote/SSE type? Remote servers are translated
  with `type` field preserved in Cursor format.
- What happens when `--platform cursor --force` is
  specified? All Cursor files are overwritten, including
  user-owned agents.
- What happens when a project has OpenSpec artifacts
  and `--platform cursor` is specified? OpenSpec
  directories are not created for Cursor (OpenSpec is
  OpenCode-specific tooling).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST accept a `--platform` flag on
  `uf init` that takes one or more platform names as a
  repeatable string slice.
- **FR-002**: System MUST support two platform values:
  `opencode` (current behavior) and `cursor` (new).
- **FR-003**: System MUST default to `--platform
  opencode` when no `--platform` flag is specified,
  preserving backward compatibility.
- **FR-004**: System MUST reject unknown platform values
  with an error message listing valid platforms.
- **FR-005**: System MUST deduplicate repeated platform
  values silently.
- **FR-006**: System MUST deploy embedded assets to
  each selected platform's directory in a single
  `Run()` invocation, iterating over platforms per
  asset file.
- **FR-007**: System MUST transform OpenCode agent
  frontmatter to Cursor format by: adding `name` from
  filename stem, preserving `description`, adding
  `model: "inherit"`, and dropping `mode`,
  `temperature`, `tools`, `maxSteps`, and `disabled`.
- **FR-008**: System MUST transform convention packs
  to Cursor `.mdc` rules by: changing file extension
  to `.mdc`, adding `description`, `globs`, and
  `alwaysApply` frontmatter fields, and prefixing the
  filename with `uf-`.
- **FR-009**: System MUST map convention pack languages
  to Cursor glob patterns: `go` to `"**/*.go"`,
  `typescript` to `"**/*.{ts,tsx,js,jsx}"`.
- **FR-010**: System MUST set `alwaysApply: true` and
  leave `globs` empty for language-agnostic packs
  (`default`, `severity`, `content` and their custom
  variants).
- **FR-011**: System MUST generate `.cursor/mcp.json`
  from detected MCP servers (Dewey, Replicator) using
  Cursor's schema: root key `mcpServers`, `command`
  as string with `args` as separate array, `env`
  instead of `environment`.
- **FR-012**: System MUST convert environment variable
  syntax from OpenCode's `{env:VAR}` format to
  Cursor's `${VAR}` format in MCP configuration.
- **FR-013**: System MUST deploy commands to
  `.cursor/commands/` (plural) as plain Markdown files
  with no content transformation.
- **FR-014**: System MUST deploy skills to
  `.cursor/skills/` with the same directory structure
  (skill-name/SKILL.md) and no content transformation.
- **FR-015**: System MUST NOT deploy OpenSpec schema
  or template assets when the Cursor platform is
  selected (OpenSpec is OpenCode-specific tooling).
- **FR-016**: System MUST apply the same file ownership
  model to Cursor files: tool-owned files are auto-
  updated on re-init, user-owned files are never
  overwritten unless `--force` is specified.
- **FR-017**: System MUST classify Cursor agents as
  user-owned and Cursor rules derived from tool-owned
  packs as tool-owned, mirroring the OpenCode ownership
  model.
- **FR-018**: System MUST support `--divisor` mode for
  the Cursor platform, deploying only Divisor agents,
  the review-council command, and language-appropriate
  convention pack rules.
- **FR-019**: System MUST NOT create `.cursorrules`
  bridge file when the Cursor platform is explicitly
  selected (native `.cursor/rules/` replaces its
  function).
- **FR-020**: System MUST continue creating `CLAUDE.md`
  and `.cursorrules` bridge files when only the
  `opencode` platform is selected (preserving current
  behavior).
- **FR-021**: System MUST create `AGENTS.md` Convention
  Packs section regardless of selected platforms, as
  it is the platform-neutral context hub.
- **FR-022**: System MUST expose a `Platform` interface
  (or equivalent abstraction) that encapsulates path
  mapping, content transformation, asset filtering,
  and MCP configuration per platform.
- **FR-023**: System MUST support the `--lang` flag for
  both platforms, controlling which language-specific
  convention packs (or `.mdc` rules) are deployed.
- **FR-024**: System MUST include `.cursor/` patterns
  in the `.gitignore` managed block when the Cursor
  platform is selected.

### Key Entities

- **Platform**: An abstraction representing a target AI
  coding tool. Each platform knows how to map asset
  paths, transform content, filter assets, and
  configure its native MCP format.
- **Asset**: An embedded file from
  `internal/scaffold/assets/` that is deployed to one
  or more platform target directories.
- **Convention Pack**: A Markdown file containing coding
  rules that is deployed as-is for OpenCode or
  transformed into a `.mdc` rule for Cursor.
- **MCP Server Entry**: A tool server configuration
  that is stored in platform-native format
  (`opencode.json` for OpenCode, `.cursor/mcp.json`
  for Cursor).

## Dependencies

- **Spec 003** (Specification Framework): Defines the
  scaffold engine and `uf init` command.
- **Spec 005** (The Divisor Architecture): Defines
  `DivisorOnly` mode, convention pack ownership, and
  the scaffold engine's `isDivisorAsset()` filter.
- **Spec 019** (Divisor Council Refinement): Defines
  severity pack (always-deploy) and convention pack
  classification.
- **Prior art**: OpenPackage `platforms.jsonc` (Cursor
  path mapping, `.mdc` extension convention) and Lola
  `CursorTarget` (agent frontmatter, MCP translation)
  informed the design.

## Assumptions

- Cursor 2.4+ supports the Agent Skills standard with
  `SKILL.md` files, matching OpenCode's skill format.
- Cursor's `.mdc` rule format is stable and uses the
  three-field frontmatter (`description`, `globs`,
  `alwaysApply`).
- Cursor's MCP format at `.cursor/mcp.json` uses the
  `mcpServers` root key with `command`/`args`/`env`
  structure (confirmed via OpenPackage and Lola).
- Cursor slash commands (`.cursor/commands/*.md`) use
  the same Markdown format as OpenCode commands
  (confirmed via Lola passthrough and OpenPackage
  direct mapping).
- The `--platform` flag follows established CLI
  conventions for repeatable string slices in Cobra
  (`StringSliceVar`).
- Future platforms (Claude Code native, Copilot, etc.)
  can be added by implementing the Platform interface
  without modifying the scaffold engine core.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can scaffold a Cursor-native
  project with a single command (`uf init --platform
  cursor`) and have all files pass Cursor's format
  validation (correct `.mdc` frontmatter, valid
  `.cursor/mcp.json` schema).
- **SC-002**: Users can scaffold a dual-platform project
  (`uf init --platform opencode --platform cursor`)
  and use both tools in the same repository without
  configuration conflicts.
- **SC-003**: Running `uf init --platform cursor`
  twice on the same project updates tool-owned files
  without overwriting user-customized agents or
  custom convention pack rules.
- **SC-004**: All existing `uf init` behavior is
  preserved when no `--platform` flag is specified
  (100% backward compatibility).
- **SC-005**: Convention packs deployed as Cursor `.mdc`
  rules use correct glob patterns that match the
  intended file types for each language.
- **SC-006**: MCP servers configured for OpenCode
  (Dewey, Replicator) are accessible in Cursor after
  running `uf init --platform cursor` in a project
  where those tools are available.
