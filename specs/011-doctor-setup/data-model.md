# Data Model: Doctor and Setup Commands

**Branch**: `011-doctor-setup` | **Date**: 2026-03-21

## Entities

### Severity

Enumeration of diagnostic result severities.

| Value | Int | Meaning |
|-------|-----|---------|
| Pass  | 0   | Check succeeded |
| Warn  | 1   | Non-critical issue or optional item missing |
| Fail  | 2   | Required item missing or invalid |

**Rules**:
- Exit code 0 when max severity across all results is Pass
  or Warn.
- Exit code 1 when any result has severity Fail.
- JSON serialization uses lowercase strings: `"pass"`,
  `"warn"`, `"fail"`.
- Optional-absent items (e.g., missing `graphthulhu`) use
  Pass severity. The text formatter renders them with the
  `○` (gray) indicator instead of `✓` (green) based on a
  convention: checks whose InstallHint is populated but
  severity is Pass are "informational" optional items.

### ManagerKind

Enumeration of detected version/package managers.

| Value     | Description |
|-----------|-------------|
| goenv     | Go version manager (shim-based) |
| nvm       | Node Version Manager (bash function, PATH-based) |
| fnm       | Fast Node Manager (binary, multishell) |
| pyenv     | Python version manager (shim-based) |
| mise      | Polyglot version manager (formerly rtx) |
| homebrew  | macOS/Linux package manager |
| bun       | Bun JavaScript runtime |
| system    | System package manager (apt, yum, etc.) |
| direct    | Official installer (e.g., Go .pkg) |
| unknown   | No manager detected |

### ManagerInfo

A detected version/package manager.

| Field    | Type        | Description |
|----------|-------------|-------------|
| Kind     | ManagerKind | Which manager |
| Path     | string      | Binary path of the manager itself |
| Manages  | []string    | Tool categories managed (e.g., "go", "node") |

### DetectedEnvironment

The developer's detected tool management landscape.

| Field    | Type          | Description |
|----------|---------------|-------------|
| Managers | []ManagerInfo | All detected managers, ordered by detection |
| Platform | string        | Runtime platform (e.g., "darwin/arm64") |

**Rules**:
- Detection runs once at startup, shared between doctor
  and setup.
- Managers are detected by binary presence in PATH and
  environment variables.
- When no managers are detected, `Managers` is empty (not
  nil).

### ToolProvenance (internal-only)

How a specific tool binary was installed. This type is used
internally during check execution to determine provenance.
It is NOT serialized in the Report JSON output. Instead,
provenance information is encoded in CheckResult.Message
(e.g., "1.24.3 via goenv") for human consumption and in
CheckResult.Detail (binary path) for programmatic access.

| Field   | Type        | Description |
|---------|-------------|-------------|
| Manager | ManagerKind | Which manager installed this tool |
| Version | string      | Detected version (e.g., "1.24.3") |
| Path    | string      | Resolved binary path |

**Rules**:
- Version is extracted from the binary's own output
  (e.g., `go version`) not from the manager's records.
- Path is the absolute binary path found in the system PATH.
- Manager is determined by path pattern analysis per R1.

### CheckResult

A single diagnostic finding.

| Field       | Type     | Required | Description |
|-------------|----------|----------|-------------|
| Name        | string   | Yes      | Check identifier (e.g., "go", "gaze", ".opencode/agents/") |
| Severity    | Severity | Yes      | Pass, Warn, or Fail |
| Message     | string   | Yes      | Human-readable result (e.g., "1.24.3 via goenv") |
| Detail      | string   | No       | Additional info (path, count, etc.) |
| InstallHint | string   | No       | Brief install command (e.g., "goenv install 1.24.3") |
| InstallURL  | string   | No       | Link to detailed docs |

**Validation rules**:
- Name MUST be non-empty.
- InstallHint MUST be present when Severity is Warn or Fail.
- InstallURL SHOULD be present when the install process is
  non-trivial.

### CheckGroup

A named collection of related check results.

| Field   | Type          | Required | Description |
|---------|---------------|----------|-------------|
| Name    | string        | Yes      | Group name (e.g., "Core Tools") |
| Results | []CheckResult | Yes      | Ordered check results |
| Embed   | string        | No       | Verbatim output from external tool (e.g., `swarm doctor`) |

**Groups** (in report order):
1. Detected Environment
2. Core Tools
3. Swarm Plugin
4. Scaffolded Files
5. Hero Availability
6. MCP Server Config
7. Agent/Skill Integrity

### Summary

Aggregate counts of check results.

| Field  | Type | Description |
|--------|------|-------------|
| Total  | int  | Total number of checks |
| Passed | int  | Count with severity Pass |
| Warned | int  | Count with severity Warn |
| Failed | int  | Count with severity Fail |

### Report

The complete diagnostic output.

| Field      | Type         | Description |
|------------|--------------|-------------|
| Groups     | []CheckGroup | Ordered check groups |
| Summary    | Summary      | Aggregate counts |
| Environment| DetectedEnvironment | Detected managers (structured, for programmatic access) |

**Note**: Environment data appears in two places:
`Report.Environment` provides structured `ManagerInfo`
objects for programmatic consumers. `Report.Groups[0]`
("Detected Environment") provides the same information as
human-readable `CheckResult` items for display. Programmatic
consumers SHOULD prefer `Report.Environment`.

## Relationships

```
Report
├── Environment (DetectedEnvironment)
│   └── Managers []ManagerInfo
│       └── Kind (ManagerKind)
├── Groups []CheckGroup
│   ├── Results []CheckResult
│   │   ├── Severity
│   │   └── (references ToolProvenance implicitly via Message)
│   └── Embed (optional, for swarm doctor)
└── Summary
```

## State Transitions

No stateful entities. Doctor produces a Report (read-only
snapshot). Setup performs idempotent install actions with no
persistent state beyond the tools it installs and the
`opencode.json` it modifies.

## JSON Serialization

Doctor JSON output serializes the Report struct with
`json:"..."` tags. Field names use `snake_case` to match
existing artifact conventions in the project.

```json
{
  "environment": {
    "managers": [
      {"kind": "goenv", "path": "/opt/homebrew/bin/goenv", "manages": ["go"]},
      {"kind": "nvm", "path": "", "manages": ["node"]},
      {"kind": "homebrew", "path": "/opt/homebrew/bin/brew", "manages": ["packages"]}
    ],
    "platform": "darwin/arm64"
  },
  "groups": [
    {
      "name": "Detected Environment",
      "results": [
        {"name": "goenv", "severity": "pass", "message": "Go version manager", "detail": "/opt/homebrew/bin/goenv"}
      ]
    },
    {
      "name": "Core Tools",
      "results": [
        {"name": "go", "severity": "pass", "message": "1.24.3 via goenv", "detail": "/Users/dev/.goenv/shims/go"},
        {"name": "opencode", "severity": "fail", "message": "not found", "install_hint": "brew install anomalyco/tap/opencode", "install_url": "https://opencode.ai/docs"}
      ]
    },
    {
      "name": "Swarm Plugin",
      "results": [
        {"name": "swarm", "severity": "pass", "message": "installed", "detail": "/usr/local/bin/swarm"}
      ],
      "embed": "✓ OpenCode plugin configured\n✓ Hive storage: libSQL (embedded SQLite)\n✓ Semantic memory: ready\n✓ Dependencies: all installed\n"
    }
  ],
  "summary": {
    "total": 20,
    "passed": 18,
    "warned": 1,
    "failed": 1
  }
}
```
