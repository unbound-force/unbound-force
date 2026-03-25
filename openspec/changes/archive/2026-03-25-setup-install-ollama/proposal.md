## Why

`uf setup` installs all required tools (OpenCode, Gaze,
Node.js, Bun, Swarm, Dewey) but does NOT install Ollama.
When Ollama is missing, `pullEmbeddingModel()` silently
returns "embedding model requires ollama" and the Granite
model is never pulled. Both Dewey (semantic search) and
Swarm (semantic memory) depend on Ollama for embedding
generation.

Currently, Ollama installation is only mentioned in a
post-setup "Tip" message. Every other required tool has
an installation step -- Ollama is the only gap.

## What Changes

Add an `installOllama()` function to `internal/setup/setup.go`
that follows the same pattern as `installOpenCode()`,
`installGaze()`, and `installDewey()`:

1. Check if `ollama` is already in PATH via `LookPath`
2. If missing and Homebrew is available, install via
   `brew install ollama`
3. If Homebrew is not available, skip with a link to
   the Ollama download page

Position the Ollama installation step **before** the
Dewey step (since `pullEmbeddingModel()` needs Ollama),
and **after** the Swarm-related steps.

Remove the post-setup "Tip" message about Ollama since
it will be installed automatically.

## Capabilities

### New Capabilities
- `setup-install-ollama`: `uf setup` installs Ollama
  via Homebrew when not present, following the same
  pattern as other tool installations.

### Modified Capabilities
- `setup-tip-message`: Remove the Ollama installation
  tip since it's now handled automatically.
- `pullEmbeddingModel`: Will now succeed on first run
  (Ollama will be present), instead of silently skipping.

### Removed Capabilities
- None

## Impact

- `internal/setup/setup.go` -- new `installOllama()`
  function, step ordering update, tip removal
- `internal/setup/setup_test.go` -- new test for Ollama
  installation, updated step order assertions

## Constitution Alignment

### I. Autonomous Collaboration
**Assessment**: N/A -- no artifact communication changes.

### II. Composability First
**Assessment**: PASS -- Ollama installation failure
produces a warning, not a hard failure. Dewey and Swarm
continue to function without Ollama (degraded mode).

### III. Observable Quality
**Assessment**: N/A -- no output format changes.

### IV. Testability
**Assessment**: PASS -- follows the existing injected
`LookPath`/`ExecCmd` pattern for testability.
