# Quickstart: Doctor & Setup Dewey Alignment

**Branch**: `023-doctor-setup-dewey` | **Date**: 2026-04-03

## What This Changes

Two targeted improvements to `uf doctor` and `uf setup`:

1. **Embedding capability check** — `uf doctor` now
   verifies that Dewey can actually generate embeddings,
   not just that the model file exists
2. **Forked Swarm plugin** — `uf setup` installs the
   Swarm plugin from the organization's fork instead of
   upstream

## Files Changed

| File | Change |
|------|--------|
| `internal/doctor/checks.go` | Add `checkEmbeddingCapability()`, update `checkDewey()` |
| `internal/doctor/doctor.go` | Add `EmbedCheck` field to `Options`, add `defaultEmbedCheck()` |
| `internal/doctor/environ.go` | Update swarm install hints to fork source |
| `internal/doctor/doctor_test.go` | Add embedding capability tests |
| `internal/setup/setup.go` | Change `installSwarmPlugin()` source to fork |
| `internal/setup/setup_test.go` | Update swarm plugin install assertions |

## How to Verify

### Embedding Capability Check

```bash
# With Ollama running and model pulled:
uf doctor
# Look for "Dewey Knowledge Layer" group:
#   ✅ dewey binary        found
#   ✅ embedding model      granite-embedding:30m installed
#   ✅ embedding capability granite-embedding:30m generating embeddings
#   ✅ workspace            initialized

# With Ollama stopped:
ollama stop  # or kill the process
uf doctor
# Look for:
#   ⚠️  embedding capability cannot generate embeddings (Ollama not running)
#        Fix: Start Ollama: ollama serve
```

### Forked Swarm Plugin

```bash
# Fresh install:
uf setup
# Look for the "Swarm plugin" step:
#   Swarm plugin...
#   ✓ Swarm plugin    installed (via bun, from unbound-force/swarm-tools)

# Verify source:
npm list -g opencode-swarm-plugin 2>/dev/null || \
bun pm ls -g 2>/dev/null | grep swarm
```

### Run Tests

```bash
# Full test suite (must pass — FR-008):
make test
# or: go test -race -count=1 ./...

# Doctor tests only:
go test -race -count=1 ./internal/doctor/...

# Setup tests only:
go test -race -count=1 ./internal/setup/...
```

## Design Decisions

1. **Ollama HTTP API over `dewey doctor`**: Direct API
   call is more reliable than parsing `dewey doctor`
   output. See research.md R1.

2. **`EmbedCheck` function injection**: Follows the
   existing `Options` pattern (`LookPath`, `ExecCmd`,
   `ReadFile`) for testability. See research.md R5.

3. **Always-install for fork**: Removed the "already
   installed" early return for swarm plugin to ensure
   the fork version is always current. See research.md R3.

4. **No new severity type**: Reuses existing `Pass`/
   `Warn`/`Fail` — no `Info` severity needed. See
   research.md R2.
