# Feature Specification: Externalize Tool Initialization

**Feature Branch**: `027-externalize-tool-init`  
**Created**: 2026-04-11  
**Status**: Draft  
**Input**: User description: "Externalize Speckit, OpenSpec, and Gaze initialization from uf embedded assets to CLI init command delegation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Speckit Initialization via specify CLI (Priority: P1)

When an engineer runs `uf init` in a project, the
Speckit framework (`.specify/` directory with scripts,
templates, and config) is created by delegating to
`specify init` rather than deploying embedded assets
from the `uf` binary. Today, `uf init` embeds 12
Speckit files (5 bash scripts, 6 Markdown templates,
1 YAML config) as scaffold assets that are deployed
once and never updated. After this change, `uf init`
calls `specify init` which creates the same structure
using the Speckit CLI's own scaffolding logic, ensuring
the files are always current with the installed version
of the `specify` CLI.

**Why this priority**: Speckit scripts and templates are
the foundation of the spec-driven development workflow.
Stale scripts cause pipeline failures (wrong flags,
missing features) and stale templates produce
inconsistent spec structures.

**Independent Test**: Install the `specify` CLI. Run
`uf init` in a fresh directory. Verify `.specify/` is
created by `specify init` (not from embedded assets).
Verify the scripts and templates match the installed
`specify` version, not the `uf` binary's embedded
version.

**Acceptance Scenarios**:

1. **Given** the `specify` CLI is installed and no
   `.specify/` directory exists, **When** the engineer
   runs `uf init`, **Then** `specify init` is called
   and creates the `.specify/` directory with scripts,
   templates, and config.
2. **Given** `.specify/` already exists, **When** the
   engineer runs `uf init`, **Then** `specify init` is
   skipped (idempotent — the directory is already
   initialized).
3. **Given** the `specify` CLI is not installed, **When**
   the engineer runs `uf init`, **Then** the Speckit
   initialization is skipped with an informational
   message and the rest of init proceeds normally.
4. **Given** `uf setup` is run on a new machine, **When**
   the setup completes, **Then** the `specify` CLI is
   installed and available for subsequent `uf init`
   runs.

---

### User Story 2 — OpenSpec Initialization via openspec CLI (Priority: P1)

When an engineer runs `uf init`, the OpenSpec framework
(`openspec/` directory with config and base structure)
is created by delegating to `openspec init --tools
opencode` rather than deploying embedded directory
scaffolding. The project-specific custom schema
(`openspec/schemas/unbound-force/`) remains embedded
in the `uf` binary and is deployed after `openspec init`
creates the base structure.

**Why this priority**: OpenSpec evolves independently
of the `uf` binary. Delegating initialization to the
OpenSpec CLI ensures new features and schema versions
are available immediately after upgrading OpenSpec,
without waiting for a new `uf` release.

**Independent Test**: Install the OpenSpec CLI. Run
`uf init` in a fresh directory. Verify `openspec/`
base structure is created by `openspec init`. Verify
the custom `unbound-force` schema is deployed on top
by `uf init`.

**Acceptance Scenarios**:

1. **Given** the OpenSpec CLI is installed and no
   `openspec/` directory exists, **When** the engineer
   runs `uf init`, **Then** `openspec init --tools
   opencode` is called to create the base structure,
   then the custom schema is deployed from embedded
   assets.
2. **Given** `openspec/` already exists, **When** the
   engineer runs `uf init`, **Then** `openspec init`
   is skipped. The custom schema is still deployed
   (tool-owned, overwritten on re-runs).
3. **Given** the OpenSpec CLI is not installed, **When**
   the engineer runs `uf init`, **Then** OpenSpec
   initialization is skipped with an informational
   message.

---

### User Story 3 — Gaze Initialization via gaze CLI (Priority: P2)

When an engineer runs `uf init`, Gaze's OpenCode
integration files (agents and commands) are created by
delegating to `gaze init` rather than expecting the
engineer to run it manually. Today, Gaze files are in
the `knownNonEmbeddedFiles` exclusion list — they exist
in the project but are not deployed by `uf init`. After
this change, `uf init` calls `gaze init` which creates
the Gaze agent and command files.

**Why this priority**: Gaze is already installed by
`uf setup` (step 2) but its per-repo initialization
is left to the engineer. This is an unnecessary manual
step that `uf init` should handle automatically.

**Independent Test**: Install Gaze. Run `uf init` in a
fresh directory. Verify `.opencode/agents/gaze-reporter.md`
and `.opencode/command/gaze.md` are created by `gaze init`.

**Acceptance Scenarios**:

1. **Given** Gaze is installed and no Gaze agent files
   exist, **When** the engineer runs `uf init`, **Then**
   `gaze init` is called and creates the Gaze agents
   and commands.
2. **Given** Gaze agent files already exist, **When**
   the engineer runs `uf init`, **Then** `gaze init` is
   skipped (idempotent).
3. **Given** Gaze is not installed, **When** the
   engineer runs `uf init`, **Then** Gaze initialization
   is skipped with an informational message.

---

### User Story 4 — Specify CLI Installation via uf setup (Priority: P1)

`uf setup` installs the `specify` CLI as a new setup
step. Today, `uf setup` does not install `specify`
because Speckit files are embedded in the `uf` binary.
After this change, `specify` is an external dependency
installed by setup, just like Gaze, Dewey, and
Replicator.

The `specify` CLI is a Python tool installed via `uv`
(the Python package manager). This adds `uv` as a new
prerequisite that `uf setup` ensures is available.

**Why this priority**: Without the `specify` CLI
installed, `uf init` cannot delegate Speckit
initialization. Setup must install it before init can
use it.

**Independent Test**: Run `uf setup` on a machine
without `specify` installed. Verify `uv` is available
after setup, and `specify` is installed and in PATH.

**Acceptance Scenarios**:

1. **Given** `uv` and `specify` are not installed,
   **When** the engineer runs `uf setup`, **Then** `uv`
   is installed (or verified), then `specify` is
   installed via `uv tool install specify-cli`.
2. **Given** `specify` is already installed, **When**
   the engineer runs `uf setup`, **Then** the specify
   step reports "already installed" and skips.
3. **Given** `uv` cannot be installed (no Python),
   **When** the engineer runs `uf setup`, **Then** the
   step fails with an actionable hint to install Python
   and uv manually.

---

### Edge Cases

- What happens when `specify init` is run inside an
  existing `.specify/` directory? The `specify` CLI
  should handle this idempotently. `uf init` checks
  for `.specify/` existence before calling.
- What happens when `openspec init` creates files that
  conflict with the embedded custom schema? `uf init`
  deploys the custom schema AFTER `openspec init`, so
  the embedded version overwrites any conflicting files
  from `openspec init`.
- What happens when the engineer upgrades `uf` but not
  `specify`? The Speckit files match the installed
  `specify` version, not the `uf` version. This is the
  desired behavior — each tool owns its own files.
- What happens when `uf init` is run without `uf setup`
  first (tools not installed)? Each tool delegation
  is gated by `LookPath` — if the binary is not in
  PATH, the step is skipped with an informational
  message. `uf init` still works for the tools that are
  available.
- What happens when `gaze init` creates files that
  `uf init` previously listed in
  `knownNonEmbeddedFiles`? The exclusion list entries
  remain valid — they prevent the
  `TestCanonicalSources_AreEmbedded` test from failing
  on Gaze-created files.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf setup` MUST install `uv` (Python
  package manager) if not already available.
- **FR-002**: `uf setup` MUST install the `specify` CLI
  via `uv tool install specify-cli` from the
  `github/spec-kit` repository.
- **FR-003**: `uf setup` MUST update the step count to
  reflect the new uv and specify steps.
- **FR-004**: `uf init` MUST delegate Speckit
  initialization to `specify init` when the `specify`
  binary is in PATH and `.specify/` does not exist.
- **FR-005**: `uf init` MUST delegate OpenSpec
  initialization to `openspec init --tools opencode`
  when the `openspec` binary is in PATH and `openspec/`
  does not exist.
- **FR-006**: `uf init` MUST delegate Gaze
  initialization to `gaze init` when the `gaze` binary
  is in PATH and Gaze agent files do not exist.
- **FR-007**: `uf init` MUST deploy the custom
  `openspec/schemas/unbound-force/` schema from
  embedded assets AFTER `openspec init` creates the
  base structure. This schema remains tool-owned
  (overwritten on re-runs).
- **FR-008**: `uf init` MUST remove the 12 Speckit
  files from embedded scaffold assets (scripts,
  templates, config).
- **FR-009**: `uf init` MUST remove the OpenSpec base
  structure files from embedded scaffold assets (only
  the custom schema remains embedded).
- **FR-010**: Each tool delegation MUST be gated by
  `LookPath` — if the binary is not in PATH, the step
  is skipped with an informational message.
- **FR-011**: Each tool delegation MUST be idempotent
  — if the tool's directory already exists, the step
  is skipped.
- **FR-012**: All existing tests MUST continue to pass
  after the changes.
- **FR-013**: The `/uf-init` slash command MUST be
  updated to document any post-init customization steps
  needed for files created by the external tool CLIs.

### Key Entities

- **Tool Delegation**: A call from `uf init` to an
  external tool's `init` command, gated by binary
  availability and directory existence.
- **Embedded Asset**: A file bundled in the `uf` binary
  and deployed during `uf init`. After this change,
  Speckit scripts/templates and OpenSpec base files are
  no longer embedded assets.
- **Custom Schema**: The project-specific
  `openspec/schemas/unbound-force/` schema that remains
  embedded in `uf` and is deployed after `openspec init`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Zero Speckit scripts or templates remain
  in `internal/scaffold/assets/specify/` — verified by
  checking the directory does not exist.
- **SC-002**: `uf init` in a fresh directory with all
  tools installed creates `.specify/`, `openspec/`, and
  Gaze agent files via CLI delegation — verified by
  running init and checking that the files match the
  installed tool versions, not embedded versions.
- **SC-003**: `uf setup` completes with the new step
  count and `specify` is available in PATH after setup
  — verified by running setup and checking `which specify`.
- **SC-004**: `uf init` in a fresh directory WITHOUT
  any tools installed still completes successfully
  (deploys OpenCode agents, packs, and commands) —
  verified by running init with no external tools.
- **SC-005**: All existing tests pass after the changes
  — verified by running the full test suite.
- **SC-006**: The custom OpenSpec schema
  (`openspec/schemas/unbound-force/`) is deployed
  correctly after `openspec init` creates the base
  structure — verified by running init and checking
  the schema files exist with correct content.

## Dependencies & Assumptions

### Dependencies

- **`specify` CLI** (`github/spec-kit`): Must be
  installable via `uv tool install`. The `specify init`
  command must create the `.specify/` directory
  structure with scripts, templates, and config.
- **`uv` Python package manager**: Required for
  installing the `specify` CLI. Must be installable
  on macOS and Fedora.
- **`openspec` CLI**: Already installed by `uf setup`
  (step 6). The `openspec init --tools opencode`
  command must create the `openspec/` base structure.
- **`gaze` CLI**: Already installed by `uf setup`
  (step 2). The `gaze init` command must create the
  Gaze agent and command files.

### Assumptions

- The `specify init` command is non-interactive when
  run without a project name argument (creates
  `.specify/` in the current directory).
- The `openspec init --tools opencode` command is
  non-interactive (the `--tools` flag bypasses the
  interactive tool selection prompt).
- The `gaze init` command is non-interactive and
  idempotent (creates files without overwriting existing
  ones unless `--force` is used).
- `uv` can be installed via `curl` on macOS and
  `dnf install` or `pip install` on Fedora. The
  installation method depends on the detected
  environment.
- The `specify` CLI creates files that are functionally
  equivalent to the currently embedded assets (same
  scripts, same templates, same config structure). If
  the `specify` CLI creates a different structure, the
  Speckit slash commands (`/speckit.specify`,
  `/speckit.plan`, etc.) may need updates.
