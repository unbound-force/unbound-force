# Feature Specification: Replicator Migration

**Feature Branch**: `024-replicator-migration`  
**Created**: 2026-04-06  
**Status**: Draft  
**Input**: User description: "issue #82 — Replace Swarm plugin with Replicator"

## User Scenarios & Testing *(mandatory)*

### User Story 1 — Replicator Installation via Setup (Priority: P1)

When an engineer runs `uf setup` on a new machine, the
setup process installs Replicator (a single Go binary)
via Homebrew instead of the Node.js Swarm plugin via
npm. Today, setup installs `opencode-swarm-plugin` via
bun or npm, requires Node.js and bun as prerequisites,
and runs `swarm setup` and `swarm init`. After this
change, setup installs `replicator` via Homebrew (same
pattern as Gaze and the `uf` binary itself), runs
`replicator setup` for per-machine initialization, and
removes the bun prerequisite entirely.

**Why this priority**: Setup is the entry point for
every new contributor. If it installs the wrong tool or
fails due to missing Node.js/bun dependencies, the
engineer cannot use the agent ecosystem at all.

**Independent Test**: Run `uf setup` on a machine
without Replicator installed. Verify Replicator is
installed via Homebrew, `replicator setup` runs
successfully, and bun is not installed or required.

**Acceptance Scenarios**:

1. **Given** Replicator is not installed, **When** the
   engineer runs `uf setup`, **Then** Replicator is
   installed via Homebrew and `replicator setup` runs
   to initialize the per-machine database and config.
2. **Given** Replicator is already installed and up to
   date, **When** the engineer runs `uf setup`, **Then**
   the Replicator install step reports "already
   installed" and `replicator setup` still runs
   (idempotent).
3. **Given** Homebrew is not available, **When** the
   engineer runs `uf setup`, **Then** the Replicator
   install step fails with an actionable hint to
   install Homebrew.
4. **Given** the engineer has the old Swarm plugin
   installed via npm, **When** they run `uf setup`,
   **Then** setup installs Replicator via Homebrew
   and does not interact with the old npm package
   (leaving cleanup to the user).

---

### User Story 2 — Project Initialization with Replicator (Priority: P1)

When an engineer runs `uf init` in a project directory,
the scaffold engine configures `opencode.json` with a
Replicator MCP server entry instead of a Swarm plugin
array entry. Today, `uf init` adds
`"opencode-swarm-plugin"` to the `plugin` array when
`.hive/` exists. After this change, `uf init` adds a
`replicator` entry to the `mcp` section when the
`replicator` binary is in PATH, and delegates to
`replicator init` for per-repo setup (creating `.hive/`).

**Why this priority**: Init is the per-repo entry point.
If `opencode.json` has the wrong configuration, agents
cannot connect to Replicator and all swarm operations
fail silently.

**Independent Test**: Run `uf init` in a fresh directory
with Replicator installed. Verify `opencode.json`
contains an `mcp.replicator` entry (not a plugin array
entry) and `.hive/` is created via `replicator init`.

**Acceptance Scenarios**:

1. **Given** Replicator is installed and no
   `opencode.json` exists, **When** the engineer runs
   `uf init`, **Then** `opencode.json` is created with
   an `mcp.replicator` entry and no `plugin` array.
2. **Given** `opencode.json` already has an
   `mcp.replicator` entry, **When** the engineer runs
   `uf init`, **Then** the entry is preserved and the
   result reports "already configured."
3. **Given** `opencode.json` has a legacy
   `opencode-swarm-plugin` in the `plugin` array,
   **When** the engineer runs `uf init`, **Then** the
   plugin entry is removed and an `mcp.replicator`
   entry is added (migration).
4. **Given** Replicator is not installed, **When** the
   engineer runs `uf init`, **Then** the Replicator
   MCP entry is skipped and no plugin array is added.
5. **Given** Replicator is installed and `.hive/` does
   not exist, **When** the engineer runs `uf init`,
   **Then** `replicator init` is called to create
   `.hive/` for per-repo setup.

---

### User Story 3 — Health Checks with Replicator (Priority: P2)

When an engineer runs `uf doctor`, the health checks
verify Replicator is installed and correctly configured
instead of checking for the Swarm plugin. Today, doctor
checks for the `swarm` binary, runs `swarm doctor`,
checks `.hive/` existence, and verifies
`opencode-swarm-plugin` is in the `opencode.json` plugin
array. After this change, doctor checks for the
`replicator` binary, delegates to `replicator doctor`,
checks `.hive/` existence, and verifies `mcp.replicator`
exists in `opencode.json`.

**Why this priority**: Doctor is diagnostic, not
blocking. An incorrect doctor check produces misleading
output but doesn't prevent work. Setup and init (US-1,
US-2) are more critical.

**Independent Test**: Run `uf doctor` with Replicator
installed and configured. Verify the output shows a
"Replicator" check group (not "Swarm Plugin") with
binary, doctor, `.hive/`, and MCP config checks.

**Acceptance Scenarios**:

1. **Given** Replicator is installed, `.hive/` exists,
   and `opencode.json` has `mcp.replicator`, **When**
   the engineer runs `uf doctor`, **Then** all
   Replicator checks report PASS.
2. **Given** Replicator is not installed, **When** the
   engineer runs `uf doctor`, **Then** the replicator
   binary check reports FAIL with install hint
   `brew install unbound-force/tap/replicator`.
3. **Given** Replicator is installed but `mcp.replicator`
   is missing from `opencode.json`, **When** the
   engineer runs `uf doctor`, **Then** the MCP config
   check reports WARN with hint `Run: uf init`.
4. **Given** Replicator is installed but
   `replicator doctor` reports issues, **When** the
   engineer runs `uf doctor`, **Then** the output
   embeds the `replicator doctor` output for diagnosis.
5. **Given** `uf doctor` is run with `--format=json`,
   **When** the output is parsed, **Then** the
   Replicator check group appears with the same
   structure as other check groups.

---

### User Story 4 — Bun Removal (Priority: P3)

Setup no longer installs bun as a prerequisite. Today,
`uf setup` runs `ensureBun()` to install bun via npm
before installing the Swarm plugin (which requires bun
at runtime). Since Replicator is a standalone Go binary,
bun is no longer needed. The OpenSpec CLI install step
switches from bun-preferred to npm-only.

**Why this priority**: Removing bun is a cleanup task
that reduces the dependency footprint. It doesn't add
new functionality.

**Independent Test**: Run `uf setup` and verify bun is
not installed or referenced. Verify the OpenSpec CLI is
installed via npm.

**Acceptance Scenarios**:

1. **Given** bun is not installed, **When** the engineer
   runs `uf setup`, **Then** bun is not installed and
   the setup step count does not include a bun step.
2. **Given** bun is already installed, **When** the
   engineer runs `uf setup`, **Then** bun is not
   referenced or checked — it is neither installed nor
   removed.
3. **Given** the OpenSpec CLI is not installed, **When**
   the engineer runs `uf setup`, **Then** the OpenSpec
   CLI is installed via npm (not bun).

---

### Edge Cases

- What happens when both the old `swarm` binary and the
  new `replicator` binary are installed? Doctor SHOULD
  check for `replicator` only and ignore the old
  `swarm` binary. Setup SHOULD not interact with the
  old binary.
- What happens when `opencode.json` has both a legacy
  `opencode-swarm-plugin` plugin entry AND an
  `mcp.replicator` entry? Init SHOULD remove the legacy
  plugin entry and preserve the MCP entry (idempotent
  migration).
- What happens when `replicator doctor` times out?
  Doctor SHOULD report WARN with "replicator doctor
  timed out" (same pattern as the current swarm doctor
  timeout handling).
- What happens when `replicator init` fails? Init
  SHOULD report the error and continue with remaining
  sub-tool initialization (non-blocking).
- What happens when `uf setup` runs without `uf init`
  afterward? Setup does per-machine work only.
  `opencode.json` configuration requires a subsequent
  `uf init` run (per-repo).
- What if the user needs to revert to the Swarm plugin?
  Rollback is manual: re-add
  `"plugin": ["opencode-swarm-plugin"]` to
  `opencode.json` and run
  `npm install -g github:unbound-force/swarm-tools`.
  Automated rollback is out of scope.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf setup` MUST install Replicator via
  Homebrew following the same pattern as Gaze
  installation (per Spec 011).
- **FR-002**: `uf setup` MUST run `replicator setup`
  after successful Replicator installation for
  per-machine initialization.
- **FR-003**: `uf setup` MUST NOT install bun or
  reference bun in any installation step.
- **FR-004**: `uf setup` MUST install the OpenSpec CLI
  via npm only (remove bun preference).
- **FR-005**: `uf setup` MUST remove the
  `installSwarmPlugin()`, `ensureBun()`,
  `runSwarmSetup()`, and `initializeHive()` functions
  and their associated setup steps.
- **FR-006**: `uf setup` MUST update the step count
  from 15 to 12 (removing bun, swarm plugin, swarm
  setup steps; replacing with replicator + replicator
  setup; removing `uf init` from setup).
- **FR-007**: `uf init` MUST add a Replicator MCP
  server entry to `opencode.json` when the `replicator`
  binary is in PATH (same detection pattern as Dewey).
- **FR-008**: `uf init` MUST remove legacy
  `opencode-swarm-plugin` entries from the `plugin`
  array in `opencode.json` when migrating to the
  Replicator MCP entry.
- **FR-009**: `uf init` MUST delegate to
  `replicator init` for per-repo setup when Replicator
  is in PATH and `.hive/` does not exist.
- **FR-010**: `uf init` MUST NOT add a `plugin` array
  entry for swarm or Replicator — the integration is
  MCP-based, not plugin-based.
- **FR-011**: `uf doctor` MUST replace the "Swarm
  Plugin" check group with a "Replicator" check group
  that verifies the `replicator` binary, delegates to
  `replicator doctor`, checks `.hive/` existence, and
  verifies `mcp.replicator` in `opencode.json`.
- **FR-012**: `uf doctor` install hints for Replicator
  MUST reference `brew install unbound-force/tap/replicator`
  (not npm or bun commands).
- **FR-013**: `uf doctor` MUST NOT reference bun in any
  install hint for any tool.
- **FR-014**: The live `opencode.json` at the repo root
  MUST be updated to replace the `plugin` array with an
  `mcp.replicator` entry.
- **FR-015**: All existing tests MUST continue to pass
  after the migration.
- **FR-016**: Install hint text in agent and command
  files that references "Swarm plugin" MUST be updated
  to reference "Replicator."

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `uf setup` completes in 12 steps (down
  from 15) with Replicator installed via Homebrew and
  zero npm/bun swarm references — verified by running
  setup and checking output.
- **SC-002**: `uf init` produces an `opencode.json` with
  `mcp.replicator` entry and zero `plugin` array
  entries — verified by running init in a fresh
  directory with Replicator installed.
- **SC-003**: `uf doctor` displays a "Replicator" check
  group with binary, doctor, `.hive/`, and MCP config
  checks — verified by running doctor with Replicator
  configured.
- **SC-004**: Zero references to `opencode-swarm-plugin`,
  `ensureBun`, `installSwarmPlugin`, or `bun add -g`
  remain in production source code — verified by text
  search.
- **SC-005**: All existing tests pass after the
  migration — verified by running the full test suite.
- **SC-006**: An engineer with the old Swarm plugin
  installed can run `uf init` and have their
  `opencode.json` automatically migrated from the
  plugin array to the MCP entry — verified by running
  init with a legacy `opencode.json`.

## Dependencies & Assumptions

### Dependencies

- **Replicator Homebrew distribution**
  (`unbound-force/replicator#2`): Complete. Replicator
  is installable via `brew install unbound-force/tap/replicator`.
- **Replicator init command**
  (`unbound-force/replicator#5`): In progress. Required
  for `uf init` to delegate per-repo setup.
- **Homebrew tap formula**
  (`unbound-force/homebrew-tap#3`): Complete. The
  formula exists in the tap.

### Assumptions

- The `replicator` binary exposes a `serve` subcommand
  that starts the MCP server on stdio (same protocol
  as Dewey's `dewey serve`). The `opencode.json` MCP
  entry uses `["replicator", "serve"]` as the command.
- The `replicator setup` command is idempotent and safe
  to run multiple times (creates config dir + database
  if missing, no-ops if present).
- The `replicator doctor` command produces text output
  suitable for embedding in `uf doctor` output (same
  pattern as the current `swarm doctor` delegation).
- The `replicator init` command creates `.hive/` in the
  current directory and is idempotent.
- OpenSpec CLI (`@fission-ai/openspec`) works correctly
  when installed via npm without bun (npm is the
  standard fallback that already works today).
- The `plugin` array in `opencode.json` may be safely
  removed when empty or when the only entry was
  `opencode-swarm-plugin` — OpenCode does not require
  a `plugin` key to exist.
