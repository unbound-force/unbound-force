# Implementation Plan: Init OpenCode Config

**Branch**: `017-init-opencode-config` | **Date**: 2026-03-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/017-init-opencode-config/spec.md`

## Summary

Move `opencode.json` management from `uf setup` to
`uf init`. Init will create/update `opencode.json` with
the Dewey MCP server entry (when dewey is available) and
the Swarm plugin entry (when `.hive/` exists). The config
is idempotent by default, with `--force` overwriting
stale entries. Setup's opencode.json step is removed
(15 total steps, down from 16). Doctor's MCP check is
fixed to use the canonical `"mcp"` key and handle
array-style command fields.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: `github.com/spf13/cobra`
(CLI), `embed.FS` (scaffold), `encoding/json` (config)
**Storage**: `opencode.json` at repo root (JSON file)
**Testing**: `go test -race -count=1 ./...` (standard
library `testing` package only)
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: CLI
**Performance Goals**: N/A (file I/O only, sub-second)
**Constraints**: Idempotent by default, `--force` for
overwrite. No new dependencies.
**Scale/Scope**: Single JSON file per repo

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration — PASS

This change modifies how `opencode.json` is written
(by init instead of setup). The MCP server entry
enables Dewey's artifact-based communication with AI
agents. No runtime coupling introduced.

### II. Composability First — PASS

Dewey MCP entry is only added when `dewey` is in PATH.
Swarm plugin is only added when `.hive/` exists. Each
tool's config is independently optional. `uf init`
works with neither, either, or both installed.

### III. Observable Quality — PASS

`uf init` reports opencode.json status as a sub-tool
result (created, configured, already configured,
skipped, failed). `uf doctor` is fixed to correctly
detect and report MCP server config status.

### IV. Testability — PASS

`ReadFile` and `WriteFile` function fields are added
to `scaffold.Options` for injectable file I/O. All
JSON config tests use `t.TempDir()` with no external
dependencies. Doctor tests use the existing injected
`ReadFile`/`LookPath` pattern.

## Project Structure

### Documentation (this feature)

```text
specs/017-init-opencode-config/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── checklists/
│   └── requirements.md  # Spec quality checklist
├── contracts/           # Phase 1 output (N/A - internal)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
internal/scaffold/
├── scaffold.go          # Add configureOpencodeJSON(),
│                        # ReadFile/WriteFile to Options,
│                        # call from initSubTools()
└── scaffold_test.go     # Tests for JSON config

internal/setup/
├── setup.go             # Remove configureOpencodeJSON(),
│                        # remove step 10, renumber to 15
└── setup_test.go        # Update tests

internal/doctor/
├── checks.go            # Fix checkMCPConfig() key and
│                        # command field extraction
└── doctor_test.go       # Update MCP check tests

cmd/unbound-force/
└── main_test.go         # Update if step count affects
                         # output assertions
```

**Structure Decision**: All changes are within existing
packages (`internal/scaffold/`, `internal/setup/`,
`internal/doctor/`). No new packages or files created.
This is a refactoring of existing functionality across
three internal packages.

## Coverage Strategy

### Unit Tests (all packages)

**scaffold package**:
- `TestConfigureOpencodeJSON_Create` — no file exists,
  dewey + hive available → creates with both entries
- `TestConfigureOpencodeJSON_DeweyOnly` — no hive dir
  → creates with mcp.dewey only, no plugin
- `TestConfigureOpencodeJSON_HiveOnly` — no dewey
  → creates with plugin only, no mcp
- `TestConfigureOpencodeJSON_Neither` — no dewey,
  no hive → no file created
- `TestConfigureOpencodeJSON_Idempotent` — both
  entries exist → unchanged, "already configured"
- `TestConfigureOpencodeJSON_AddMissing` — plugin
  exists but no mcp.dewey → adds mcp.dewey
- `TestConfigureOpencodeJSON_PreserveCustom` — custom
  MCP servers → preserved alongside dewey
- `TestConfigureOpencodeJSON_Force` — stale mcp.dewey
  → overwritten with correct config
- `TestConfigureOpencodeJSON_Malformed` — invalid JSON
  → skipped with warning
- `TestConfigureOpencodeJSON_ReadOnly` — write fails
  → reports failure, non-fatal

**setup package**:
- Update `TestSetupRun_OpencodeJsonManipulation` —
  verify opencode.json is NOT written by setup
- Update `TestSetupRun_NoOpencodeJson` — verify setup
  doesn't create it
- Update step count assertions (16 → 15)

**doctor package**:
- `TestCheckMCPConfig_MpcKey` — uses `"mcp"` key
  → finds servers
- `TestCheckMCPConfig_McpServersKey` — uses legacy
  `"mcpServers"` key → finds servers (fallback)
- `TestCheckMCPConfig_ArrayCommand` — command is
  `["dewey", "serve"]` → extracts `dewey` as binary
- `TestCheckMCPConfig_StringCommand` — command is
  `"dewey"` → extracts `dewey` as binary (backward
  compat)

### Coverage Targets

- Scaffold `configureOpencodeJSON()`: 100% branch
  coverage (all 10 test cases above)
- Doctor `checkMCPConfig()`: 100% branch coverage
  for key detection and command extraction
- Global coverage: maintain ≥ 80% threshold

## Complexity Tracking

No constitution violations. No complexity justifications
needed.
