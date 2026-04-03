# Data Model: Doctor & Setup Dewey Alignment

**Branch**: `023-doctor-setup-dewey` | **Date**: 2026-04-03

## Overview

This feature modifies two existing Go packages. No new
data types are introduced. Changes are limited to:
1. A new function field on `doctor.Options`
2. Modified check logic in `checks.go`
3. Modified install source in `setup.go`

## Modified Types

### `doctor.Options` (internal/doctor/doctor.go)

New injectable function field for embedding capability
verification:

```go
type Options struct {
    // ... existing fields ...

    // EmbedCheck tests whether the embedding model can
    // generate embeddings. Returns nil on success or an
    // error describing the failure. Injected for
    // testability per Constitution Principle IV.
    //
    // Production implementation sends a POST request to
    // the Ollama /api/embed endpoint with a minimal test
    // input. The endpoint URL is derived from OLLAMA_HOST
    // env var (default: http://localhost:11434).
    EmbedCheck func(model string) error
}
```

**Default implementation** (in `defaults()`):

```go
if o.EmbedCheck == nil {
    o.EmbedCheck = defaultEmbedCheck(o.Getenv)
}
```

The `defaultEmbedCheck` function:
1. Reads `OLLAMA_HOST` from environment (default
   `http://localhost:11434`)
2. POSTs to `{host}/api/embed` with body
   `{"model": model, "input": "test"}`
3. Parses the JSON response
4. Returns `nil` if `embeddings` array is non-empty
5. Returns descriptive error otherwise

### `setup.Options` (internal/setup/setup.go)

No changes to the `Options` struct. The
`installSwarmPlugin()` function changes its install
source string only.

## Existing Types (Unchanged)

### `doctor.Severity` (internal/doctor/models.go)

No changes. The existing `Pass`, `Warn`, `Fail`
severities are sufficient:
- Embedding check PASS → `Pass`
- Embedding check FAIL → `Warn` (Dewey is optional per
  Constitution Principle II — Composability First)
- Ollama demotion → `Pass` with informational message

### `doctor.CheckResult` (internal/doctor/models.go)

No changes. All new checks use the existing fields:
- `Name`: check identifier
- `Severity`: pass/warn/fail
- `Message`: human-readable status
- `InstallHint`: actionable fix command
- `Detail`: additional context (path, etc.)

### `doctor.CheckGroup` (internal/doctor/models.go)

No changes. The "Dewey Knowledge Layer" group already
exists in `checkDewey()`.

## Data Flow

### Embedding Capability Check

```text
uf doctor
  → Run()
    → checkDewey(opts)
      → LookPath("dewey")     [existing: binary check]
      → ExecCmd("ollama","list") [existing: model check]
      → EmbedCheck("granite-embedding:30m")  [NEW]
        → POST http://localhost:11434/api/embed
        → Parse JSON response
        → Return nil or error
      → CheckResult{Name:"embedding capability", ...}
```

### Swarm Plugin Installation

```text
uf setup
  → Run()
    → installSwarmPlugin(opts, env)
      → ExecCmd("bun","add","-g",
          "github:unbound-force/swarm-tools")  [CHANGED]
        OR
      → ExecCmd("npm","install","-g",
          "github:unbound-force/swarm-tools")  [CHANGED]
```

## JSON Output Schema

No schema changes. The embedding capability check
produces a standard `CheckResult` that serializes
identically to existing checks:

```json
{
  "name": "embedding capability",
  "severity": "pass",
  "message": "granite-embedding:30m generating embeddings"
}
```

Or on failure:

```json
{
  "name": "embedding capability",
  "severity": "warn",
  "message": "cannot generate embeddings",
  "install_hint": "Start Ollama: ollama serve, then: ollama pull granite-embedding:30m"
}
```

## Dependency Graph

```text
doctor.Options.EmbedCheck
  ├── uses: doctor.Options.Getenv (for OLLAMA_HOST)
  └── uses: net/http (standard library, no new deps)

setup.installSwarmPlugin
  └── uses: setup.Options.ExecCmd (existing)
```

No new external dependencies. Only `net/http` from the
Go standard library is added to `internal/doctor/`.
