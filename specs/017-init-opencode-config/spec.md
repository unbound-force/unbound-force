---
spec_id: "017"
title: "Init OpenCode Config"
status: draft
created: 2026-03-29
branch: 017-init-opencode-config
phase: 3
depends_on:
  - "[[specs/003-specification-framework/spec]]"
  - "[[specs/011-doctor-setup/spec]]"
  - "[[specs/015-dewey-integration/spec]]"
---

# Feature Specification: Init OpenCode Config

**Feature Branch**: `017-init-opencode-config`
**Created**: 2026-03-29
**Status**: Draft
**Input**: Move opencode.json management from uf setup
to uf init with Dewey MCP server config, Swarm plugin
config, idempotent behavior, and force flag support.

## User Scenarios & Testing *(mandatory)*

### User Story 1 -- Fresh Repo Init (Priority: P1)

A developer runs `uf init` in a new repo that has no
`opencode.json`. The command creates the file with both
the Swarm plugin entry and the Dewey MCP server entry
(when the respective tools are installed), so the
developer's AI agents immediately have access to Dewey's
semantic search and Swarm's orchestration tools.

**Why this priority**: Without this, the developer must
manually create `opencode.json` and add the MCP server
config -- the most common source of "why doesn't Dewey
work?" confusion.

**Independent Test**: Run `uf init` in an empty temp
directory with dewey and swarm available, then verify
`opencode.json` contains both the `mcp.dewey` entry
and the `plugin` array.

**Acceptance Scenarios**:

1. **Given** a directory with no `opencode.json` and
   `dewey` is in PATH,
   **When** the user runs `uf init`,
   **Then** `opencode.json` is created with the
   `$schema`, the `mcp.dewey` server entry
   (`type: local`, command `dewey serve --vault .`,
   `enabled: true`), and the `plugin` array containing
   `opencode-swarm-plugin`.

2. **Given** a directory with no `opencode.json` and
   `dewey` is NOT in PATH but the swarm plugin is
   available,
   **When** the user runs `uf init`,
   **Then** `opencode.json` is created with the
   `$schema` and the `plugin` array, but no `mcp`
   section.

3. **Given** a directory with no `opencode.json` and
   neither `dewey` nor the swarm plugin are available,
   **When** the user runs `uf init`,
   **Then** no `opencode.json` is created.

---

### User Story 2 -- Idempotent Re-run (Priority: P1)

A developer runs `uf init` again in a repo that already
has an `opencode.json` with existing configuration. The
command adds missing entries without disturbing existing
config. If both entries are already present, it skips
with no changes.

**Why this priority**: Idempotency is critical because
`uf init` is designed to be run multiple times safely.
Overwriting user-customized MCP server config would
break their setup.

**Independent Test**: Create an `opencode.json` with
custom MCP servers, run `uf init`, verify the custom
entries are preserved and the Dewey/Swarm entries are
added.

**Acceptance Scenarios**:

1. **Given** `opencode.json` already has both
   `mcp.dewey` and `plugin` entries,
   **When** the user runs `uf init`,
   **Then** `opencode.json` is unchanged and the
   summary reports "already configured".

2. **Given** `opencode.json` has a `plugin` entry but
   no `mcp.dewey`,
   **When** the user runs `uf init` with `dewey` in
   PATH,
   **Then** the `mcp.dewey` entry is added and the
   existing `plugin` array is preserved.

3. **Given** `opencode.json` has custom MCP servers
   (e.g., `mcp.my-custom-server`),
   **When** the user runs `uf init`,
   **Then** the custom MCP server entries are preserved
   and `mcp.dewey` is added alongside them.

---

### User Story 3 -- Force Overwrite (Priority: P2)

A developer has a stale or broken Dewey MCP config
(e.g., with the old `--include-hidden` flag). Running
`uf init --force` overwrites the Dewey MCP entry with
the current correct config.

**Why this priority**: Force overwrite is needed to
fix known stale config issues, but it's less common
than the initial setup or idempotent re-run scenarios.

**Independent Test**: Create an `opencode.json` with a
stale `mcp.dewey` entry, run `uf init --force`, verify
the entry is replaced with the current correct config.

**Acceptance Scenarios**:

1. **Given** `opencode.json` has `mcp.dewey` with
   stale command args
   `["dewey", "serve", "--include-hidden", "--vault", "."]`,
   **When** the user runs `uf init --force`,
   **Then** the `mcp.dewey` command is replaced with
   `["dewey", "serve", "--vault", "."]`.

2. **Given** `opencode.json` has `mcp.dewey` with
   the correct config,
   **When** the user runs `uf init --force`,
   **Then** `mcp.dewey` is overwritten with the same
   content and the summary reports "overwritten".

---

### User Story 4 -- Setup Delegates to Init (Priority: P2)

`uf setup` no longer directly writes to `opencode.json`.
Since setup's final step runs `uf init`, the Swarm
plugin and Dewey MCP config are still configured during
setup -- just indirectly via init.

**Why this priority**: This is a refactoring concern
that simplifies ownership. The user-visible behavior
of `uf setup` is unchanged, but the responsibility
moves to init.

**Independent Test**: Run `uf setup` in a temp
directory and verify `opencode.json` is created by
init (at step 16), not by setup directly.

**Acceptance Scenarios**:

1. **Given** a fresh environment with dewey and swarm
   available,
   **When** the user runs `uf setup`,
   **Then** `opencode.json` is created with both the
   `mcp.dewey` entry and the `plugin` array (created
   by `uf init` at the final step).

2. **Given** `uf setup` is run,
   **When** the step sequence is displayed,
   **Then** the total step count is 15 (not 16) and
   there is no opencode.json step -- the config is
   handled transparently by `uf init` at the final
   step.

---

### Edge Cases

- What happens when `opencode.json` contains malformed
  JSON? The command skips the file with a warning and
  does not attempt to modify it.
- What happens when `opencode.json` is read-only? The
  command reports a write failure and continues
  (non-fatal).
- What happens when only `dewey` is available but not
  the swarm plugin? Only the `mcp.dewey` entry is
  added, no `plugin` array.
- What happens when only the swarm plugin is available
  but not `dewey`? Only the `plugin` array is added,
  no `mcp` section.
- What happens when `opencode.json` uses the legacy
  `"mcpServers"` key? `uf doctor` reads both keys;
  `uf init` writes the canonical `"mcp"` key only.
  If `mcpServers.dewey` exists, init treats it as
  "already configured" and does not add a duplicate
  `mcp.dewey` entry.
- What happens during `uf setup --dry-run`? The
  `runUnboundInit()` function returns before calling
  `scaffold.Run()`, and scaffold's own `DryRun` guard
  prevents `configureOpencodeJSON()` from writing.
  No dry-run leak.
- What happens when `uf init --divisor` is run?
  `opencode.json` is not created or modified
  (`DivisorOnly` mode skips sub-tool initialization).
- What happens when two `uf init` processes run
  concurrently? Last writer wins. This is acceptable
  because `uf init` is a developer-facing CLI tool,
  not a server process.
- What happens when `ReadFile` fails with a non-"not
  found" error (e.g., permission denied on an existing
  file)? The command returns action `"error"` with a
  detail describing the read failure.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `uf init` MUST add the Dewey MCP server
  entry to `opencode.json` when `dewey` is available
  in PATH.
- **FR-002**: `uf init` MUST add the Swarm plugin
  entry (`opencode-swarm-plugin`) to `opencode.json`
  when the swarm plugin is available.
- **FR-003**: `uf init` MUST create `opencode.json`
  with a `$schema` field when the file does not exist
  and either dewey or swarm is available.
- **FR-004**: `uf init` MUST NOT create `opencode.json`
  when neither dewey nor swarm is available.
- **FR-005**: `uf init` MUST preserve existing entries
  in `opencode.json` (custom MCP servers, custom
  config) when adding new entries.
- **FR-006**: `uf init` MUST be idempotent -- running
  it multiple times with the same config produces no
  changes after the first run.
- **FR-007**: `uf init --force` MUST overwrite the
  `mcp.dewey` entry even if it already exists, to fix
  stale config.
- **FR-008**: `uf init` MUST skip `opencode.json`
  modification when the file contains malformed JSON,
  reporting a warning.
- **FR-009**: `uf setup` MUST NOT directly write to
  `opencode.json`. The opencode.json step in setup
  MUST be removed or replaced with a delegation note.
- **FR-010**: `uf init` MUST report the opencode.json
  configuration status in its sub-tool results
  (created, configured, already configured, overwritten,
  skipped, error, or failed).
- **FR-011**: The `Options` struct in `scaffold.go`
  MUST include `ReadFile` and `WriteFile` function
  fields for injectable file I/O, enabling test
  isolation.
- **FR-012**: `uf doctor` MUST check for MCP server
  entries using both the `"mcp"` key (canonical) and
  the `"mcpServers"` key (legacy fallback) in
  `opencode.json`, fixing the current key mismatch
  bug.
- **FR-013**: `uf doctor` MUST verify the Dewey MCP
  server binary is available when `mcp.dewey` is
  configured in `opencode.json`.
- **FR-014**: `uf doctor` MUST correctly extract the
  binary name from both string-style (`"command":
  "dewey"`) and array-style (`"command": ["dewey",
  "serve", "--vault", "."]`) MCP server command
  fields.
- **FR-015**: `uf init` MUST treat an existing
  `mcpServers.dewey` entry (legacy key) as "already
  configured" and MUST NOT add a duplicate
  `mcp.dewey` entry alongside it.
- **FR-016**: `uf init` MUST produce byte-identical
  `opencode.json` output when no logical changes are
  made, ensuring idempotent re-runs do not create
  git noise.
- **FR-017**: The `Options` struct in `scaffold.go`
  MUST include a `DryRun` boolean field. When
  `DryRun` is true, `configureOpencodeJSON()` MUST
  NOT write any file and MUST return a "dry-run"
  action result.
- **FR-018**: `uf setup` MUST forward `ReadFile`,
  `WriteFile`, and `DryRun` from its Options to
  `scaffold.Options` in `runUnboundInit()`, ensuring
  the injection chain is not broken.
- **FR-019**: `configureOpencodeJSON()` MUST return
  distinct action values for distinct conditions:
  `"error"` for malformed JSON or read failures,
  `"skipped"` for nothing to configure. These MUST
  NOT be conflated.
- **FR-020**: `configureOpencodeJSON()` MUST update
  the `subToolResult` struct comment to include all
  action values in its vocabulary.
- **FR-021**: `printSummary()` MUST render new action
  values appropriately: `"created"`, `"configured"`,
  `"already configured"`, `"overwritten"` display
  with `✓`; `"skipped"` displays with `—`;
  `"error"` and `"failed"` display with `✗`.

### Key Entities

- **opencode.json**: OpenCode configuration file at
  repo root. Contains `$schema`, `mcp` (MCP server
  definitions), and `plugin` (installed plugins).
- **MCP server entry**: A named server config under
  the `mcp` key with `type`, `command`, and `enabled`
  fields.
- **Plugin entry**: A string in the `plugin` array
  identifying an installed OpenCode plugin.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: After `uf init` in a fresh repo with
  dewey installed, `opencode.json` contains a valid
  `mcp.dewey` entry with `type: local`,
  `command: ["dewey", "serve", "--vault", "."]`, and
  `enabled: true`.
- **SC-002**: Running `uf init` twice produces
  byte-identical `opencode.json` content -- no
  duplicate entries, no lost config, no key reordering.
- **SC-003**: Running `uf init --force` replaces a
  stale `mcp.dewey` entry with the current correct
  config.
- **SC-004**: Running `uf setup` still results in a
  configured `opencode.json` (via init delegation),
  with no behavioral regression for users.
- **SC-005**: All existing tests pass after the
  refactoring, including updated setup tests that no
  longer test direct opencode.json manipulation.
- **SC-006**: Custom MCP server entries in
  `opencode.json` are preserved across `uf init`
  runs -- zero data loss.

## Clarifications

### Session 2026-03-29

- Q: How should swarm plugin availability be detected
  in `uf init`? → A: Use `.hive/` directory existence
  as proxy for swarm availability.
- Q: How should setup handle the removed opencode.json
  step? → A: Remove the step entirely and renumber
  from 16 to 15 total steps.
- Q: Which key format should be used for MCP server
  entries in opencode.json? → A: Use `"mcp"` as the
  canonical key (matches current repo).
- Q: Should `uf doctor` be updated to fix the MCP key
  mismatch (checks `"mcpServers"` but file uses
  `"mcp"`)? → A: Yes, check both keys (`"mcp"` first,
  `"mcpServers"` fallback).
- Q: Should doctor fix the binary extraction for
  array-style command fields? → A: Yes, extract first
  element of array as binary name.
- Q: Should ReadFile/WriteFile be forwarded from
  runUnboundInit() to scaffold Options? → A: Yes, add
  forwarding task to maintain injection chain.
- Q: Should subToolResult action vocabulary be updated
  and printSummary() rendering specified? → A: Yes,
  add tasks for struct comment and rendering spec.
- Q: Should "skipped" be split for distinct conditions?
  → A: Yes, use "error" for malformed JSON / read
  failures, "skipped" for nothing to configure.
- Q: Should legacy mcpServers.dewey be treated as
  already configured? → A: Yes, skip adding mcp.dewey
  if mcpServers.dewey exists.
- Q: Should byte-identical output be required for
  idempotent re-runs? → A: Yes, add requirement and
  test.
- Q: Should scaffold Options include a DryRun field?
  → A: Yes, add DryRun boolean for explicit dry-run
  guard in configureOpencodeJSON().

## Assumptions

- The `opencode.json` MCP server entry format uses the
  canonical key `mcp` (not `mcpServers`). This matches
  the current repo's file structure and is the format
  OpenCode reads.
- The Dewey MCP server command is
  `["dewey", "serve", "--vault", "."]` with no
  additional flags.
- The Swarm plugin availability is detected by checking
  if the `.hive/` directory exists in the target
  directory, which serves as a proxy for swarm
  availability (created by `swarm init` during
  `initSubTools()`).
- The `--force` flag already exists on `uf init` and
  flows through to `scaffold.Options.Force`.
- Setup step 16 (`runUnboundInit`) already calls
  `scaffold.Run()` internally, so init's new
  opencode.json management will execute during setup.
- This spec supersedes Spec 011 FR-027, FR-027a, and
  FR-028. Those requirements are now fulfilled by
  `uf init` instead of `uf setup`. Spec 011's
  implementation in `setup.go` will be removed as
  part of US4.
- The `.hive/` directory check uses `os.Stat` directly
  (consistent with existing `initSubTools()` patterns).
  Tests create the directory in `t.TempDir()` rather
  than injecting a stat function.
- `configureOpencodeJSON()` MUST run after all
  sub-tool initialization steps that create `.hive/`
  within `initSubTools()`.
- `opencode.json` is generated dynamically by
  `configureOpencodeJSON()` based on tool availability,
  not deployed as a static scaffold asset.
- Tests MUST parse JSON output and assert individual
  fields, not compare raw JSON strings (key ordering
  is deterministic via `json.MarshalIndent` on maps
  but tests should be resilient).
