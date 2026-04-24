---
status: draft
branch: opsx/unified-config
date: 2026-04-23
---

## Why

The Unbound Force CLI (`uf`) makes opinionated decisions about package
managers (Homebrew), container runtimes (Podman), embedding models
(IBM Granite via Ollama), agent LLM models (Claude Opus via Vertex),
sandbox resource limits, gateway ports, and doctor check severities.
These defaults are hardcoded as Go constants across `internal/setup/`,
`internal/scaffold/`, `internal/doctor/`, `internal/sandbox/`, and
`internal/gateway/` -- unreachable without code changes.

Today there are 6 separate configuration surfaces (`opencode.json`,
`.specify/config.yaml`, `openspec/config.yaml`, `.uf/sandbox.yaml`,
`.uf/config.yaml`, `.golangci.yml`) with no unified override
mechanism. Users on Fedora (dnf), Debian (apt), or those preferring
Docker over Podman, or remote embedding services over local Ollama,
must either fork the code or accept the defaults.

A standardized, layered configuration file would:

- **Reduce adoption friction**: Users customize defaults via a single
  file instead of editing Go constants or agent frontmatter.
- **Enable CI/CD integration**: Teams commit project-level config
  (skip Ollama in CI, use a specific sandbox image), while developers
  keep personal preferences locally.
- **Consolidate config surfaces**: Absorb `.uf/sandbox.yaml` and the
  existing (but never consumed) `.uf/config.yaml` workflow section.
- **Wire dormant infrastructure**: The existing `installViaRpm()`
  function in `setup.go` and `ManagerDnf` detection in `environ.go`
  are implemented but never called. Config makes them reachable.

## What Changes

Introduce a unified `.uf/config.yaml` configuration file with layered
resolution (CLI flags > env vars > repo config > user config >
compiled defaults), a new `internal/config/` package, a `uf config`
command group (init, show, validate), and integration across all
CLI subsystems.

Separately, remove hardcoded `model:` fields from agent frontmatter
to let OpenCode's native model resolution handle LLM selection --
this is already supported by OpenCode's config hierarchy and does
not require UF-level configuration.

## Capabilities

### New Capabilities

- `uf config init`: Creates or updates `.uf/config.yaml` with a
  fully commented-out template. When re-run on an existing file,
  preserves user-uncommented values, adds new sections from the
  current UF version, and removes deprecated sections -- ensuring
  the file is always schema-valid after the command.
- `uf config show`: Displays the effective configuration after all
  layers are merged (user + repo + env vars + defaults). Supports
  `--format json` for machine consumption.
- `uf config validate`: Validates `.uf/config.yaml` against a JSON
  Schema, reporting all errors doctor-style (reuses existing
  `internal/schemas/` validation infrastructure).
- `internal/config/` package: `Config` struct, `Load()` with
  layered merge, `InitFile()` with comment-preserving update,
  `Validate()` with JSON Schema, `Template()` for the full
  commented-out YAML.
- User-level config at `~/.config/uf/config.yaml`: Personal
  preferences (package manager, tool methods) that apply across
  all repositories. Repo-level config overrides user-level on
  conflict.
- Per-tool install method override in `setup.tools.<name>.method`:
  Allows specifying `homebrew`, `dnf`, `apt`, `rpm`, `curl`,
  `skip`, or `auto` per tool, wiring the existing dormant
  `installViaRpm()` infrastructure.
- Doctor "Configuration" check group: Validates config files exist,
  are schema-valid, and reports sandbox.yaml deprecation.
- CI/CD pattern: `setup.skip` to exclude tools, `setup.package_manager:
  manual` to skip all auto-install, `doctor.skip` to suppress checks.

### Modified Capabilities

- `uf init`: No longer creates `.uf/config.yaml` (moved to
  `uf config init`). The existing `workflowConfigContent` constant
  in `scaffold.go` is removed.
- `uf setup`: Reads `setup.*` config section. Respects
  `package_manager`, `skip`, and per-tool `method` overrides.
  Existing auto-detect behavior is the default when no config exists.
- `uf doctor`: Reads `doctor.*` config section. New "Configuration"
  check group. Respects `skip` list and tool severity overrides.
- `uf sandbox`: Reads `sandbox.*` from unified config. Falls back
  to `.uf/sandbox.yaml` if sandbox section is absent (backward
  compatibility). Prints deprecation warning when sandbox.yaml found.
- `uf gateway`: Reads `gateway.*` config section for default port
  and provider.
- Agent frontmatter: `model:` field removed from 12 agent files.
  Agents inherit the model from OpenCode's own config hierarchy
  (project `opencode.json` > user `~/.config/opencode/opencode.json`).

### Removed Capabilities

- `.uf/sandbox.yaml` (deprecated): Absorbed into `.uf/config.yaml`
  `sandbox:` section. Backward-compatible fallback reads sandbox.yaml
  when the unified config has no sandbox section. Future major version
  removes the fallback entirely.
- `workflowConfigContent` constant: The scaffold engine no longer
  creates `.uf/config.yaml` during `uf init`. Config file creation
  is the responsibility of `uf config init`.

## Impact

### Files Created (new)

- `internal/config/config.go` -- Config struct, Load(), merge logic
- `internal/config/defaults.go` -- Compiled defaults function
- `internal/config/template.go` -- Full commented-out YAML template
- `internal/config/init.go` -- InitFile() create/update logic
- `internal/config/schema.go` -- JSON Schema definition + Validate()
- `internal/config/config_test.go` -- Tests
- `cmd/unbound-force/config.go` -- uf config command group

### Files Modified

- `cmd/unbound-force/main.go` -- Register newConfigCmd()
- `cmd/unbound-force/sandbox.go` -- Load config, pass sandbox section
- `cmd/unbound-force/gateway.go` -- Load config, pass gateway section
- `internal/setup/setup.go` -- Add config fields to Options struct
- `internal/scaffold/scaffold.go` -- Remove workflowConfigContent,
  stop creating .uf/config.yaml in uf init
- `internal/doctor/doctor.go` -- Add Configuration check group
- `internal/doctor/checks.go` -- Config validation check, read
  embedding config for model name
- `internal/sandbox/workspace.go` -- Fallback to sandbox.yaml with
  deprecation warning
- `internal/sandbox/backend.go` -- Read runtime config (podman/docker)
- `.opencode/agents/*.md` -- Remove model: from 12 agent files
- `internal/scaffold/assets/opencode/agents/*.md` -- Sync scaffold copies

### Backward Compatibility

- No config file required. Missing file = compiled defaults. Zero
  friction for existing users.
- `.uf/sandbox.yaml` continues to work with deprecation warning.
- All existing CLI flags and environment variables keep their
  current behavior and precedence.
- Existing `workflow:` section in `.uf/config.yaml` (if manually
  created) is preserved -- it becomes one section among seven.

## Documentation Impact

This change introduces 3 new CLI commands and modifies 5
existing commands. The following documentation updates are
required:

- **AGENTS.md**: Update Project Structure (add `internal/config/`,
  `cmd/unbound-force/config.go`), Active Technologies, and
  Recent Changes. Covered by tasks 11.1-11.3.
- **Website**: A documentation issue MUST be filed in
  `unbound-force/website` covering the `uf config` command
  group, layered resolution model, and CI/CD configuration
  patterns. To be filed during implementation.
- **Blog opportunity**: The unified config feature reduces
  config surfaces and enables CI/CD patterns — suitable for
  a blog post. Issue to be filed when the feature is closer
  to completion.

## Constitution Alignment

Assessed against the Unbound Force org constitution.

### I. Autonomous Collaboration

**Assessment**: PASS

Configuration is a local per-repo file. No inter-hero runtime
coupling is introduced. Heroes continue to collaborate through
artifacts. The config file itself is a self-describing artifact
(YAML with JSON Schema validation, versioned template). Config
loading is a cmd-layer concern that populates Options structs --
internal packages remain unaware of config file mechanics.

### II. Composability First

**Assessment**: PASS

This change strengthens composability. The `scaffold.heroes`
section allows users to deploy only the heroes they need. The
`setup.skip` list lets users exclude tools they don't use. The
config file is optional -- its absence produces identical behavior
to today. No hero requires the config file to function. Each
subsystem's defaults remain self-sufficient.

### III. Observable Quality

**Assessment**: PASS

`uf config show` produces machine-parseable JSON output (effective
config after merge). `uf config validate` produces structured
validation results (reusing the doctor CheckResult/CheckGroup
model). The config file itself is validated against a JSON Schema.
Doctor gains a "Configuration" check group that reports config
health in its standard pass/warn/fail format.

### IV. Testability

**Assessment**: PASS

The `internal/config/` package follows the established injectable
function pattern (`ReadFile`, `Getenv` on a LoadOptions struct).
Config loading, merging, template generation, and schema validation
are all independently testable with no filesystem or environment
dependencies. The existing `internal/schemas/ValidateBytes()`
function is reused for config validation. The cmd layer's config
wiring is testable through the existing params struct pattern.
