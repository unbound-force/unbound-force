# Implementation Plan: Dewey Unified Memory

**Branch**: `021-dewey-unified-memory` | **Date**: 2026-04-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/021-dewey-unified-memory/spec.md`

## Summary

Unify the swarm's semantic memory layer under Dewey by:
(1) adding Ollama lifecycle management to `dewey serve`
so embeddings are always available without manual setup,
(2) adding learning storage to Dewey so agents can
persist and retrieve session learnings through the same
search interface used for specs and code,
(3) migrating all agent files in this repo from Hivemind
tools to Dewey equivalents, and (4) forking the Swarm
plugin to `unbound-force/swarm` for full control over
Hivemind deprecation.

This is a **cross-repo spec** — changes span the Dewey
repo (Go), this meta repo (Markdown + Go tests), and
the Swarm fork (Node.js/TypeScript). Implementation is
phased: Dewey first, then this repo, then Swarm fork.
The plan documents what needs to happen in each repo
without implementing cross-repo changes directly.

## Technical Context

**Language/Version**: Go 1.24+ (Dewey repo: Ollama
lifecycle, learning storage), Markdown (this repo:
agent files, commands), Go 1.24+ (this repo: doctor/
setup/scaffold Go code), Node.js/TypeScript (Swarm
fork: Hivemind modifications)
**Primary Dependencies**: Dewey MCP server, Ollama HTTP
API (`localhost:11434`), `github.com/spf13/cobra` (CLI),
`embed.FS` (scaffold assets), `github.com/charmbracelet/log`
**Storage**: Dewey `graph.db` (SQLite — learnings
persisted here), `.hive/` (Swarm state, unchanged)
**Testing**: Standard library `testing` package;
`go test -race -count=1 ./...` for this repo. Dewey
repo has its own test suite.
**Target Platform**: macOS, Linux (CI: ubuntu-latest)
**Project Type**: Cross-repo specification (meta repo
+ Dewey repo + Swarm fork)
**Performance Goals**: Ollama health check < 500ms,
learning storage < 100ms, learning search via
`dewey_semantic_search` < 200ms
**Constraints**: Backward compatible — existing
workflows must not break during migration. Graceful
degradation preserved for all Dewey-dependent features.
No Hivemind data migration (learnings abandoned).
**Scale/Scope**: 3 repos affected, 21 FRs, ~15
Markdown files modified in this repo, ~5 Go files
modified in this repo, new Go package in Dewey repo

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check
after Phase 1 design.*

### I. Autonomous Collaboration — PASS

Heroes communicate through well-defined artifacts.
This spec changes the *storage backend* for learnings
(Hivemind → Dewey) but preserves the artifact-based
communication pattern. Learnings are stored as documents
in Dewey's index — they are self-describing with
`source_type: "learning"` provenance metadata. No hero
depends on another hero's learning storage call. The
Ollama lifecycle management is internal to Dewey — no
inter-hero runtime coupling.

### II. Composability First — PASS

**Key principle for this spec.** Three critical
composability checks:

1. **Dewey without Ollama**: Dewey operates in
   keyword-only mode when Ollama is unavailable
   (FR-006). No hard dependency on Ollama.
2. **Agents without Dewey**: All learning storage and
   retrieval steps gracefully degrade when Dewey is
   unavailable (FR-014). The 3-tier fallback pattern
   (Full Dewey, Graph-only, No Dewey) is preserved.
3. **Swarm without Dewey**: The forked Swarm plugin
   functions identically for all non-Hivemind tools
   (FR-017). Hivemind tools proxy through Dewey or
   degrade gracefully.

The Swarm fork is the most sensitive composability
concern. The fork MUST NOT break any existing Swarm
tool. FR-017 explicitly requires identical behavior
for non-Hivemind tools.

### III. Observable Quality — PASS

Learnings stored in Dewey include provenance metadata
(`source_type: "learning"`, `source_id`, `created_at`,
tags). Search results distinguish learnings from specs,
code, and web documentation. The `dewey_store_learning`
tool returns a structured response confirming storage
and indexing status. Doctor health checks produce
machine-parseable output (JSON format) for embedding
capability status.

### IV. Testability — PASS

**Dewey repo**: Ollama lifecycle management is testable
via dependency injection (HTTP client for health checks,
subprocess launcher for `ollama serve`). Learning
storage is testable against an in-memory or temp-file
SQLite database. No external services required.

**This repo**: Agent file changes are verified by
existing scaffold drift detection tests. Go code
changes (doctor/setup) use injected `LookPath`,
`ExecCmd`, `ReadFile` functions — all testable with
`t.TempDir()` and mock functions. No network access.

**Swarm fork**: Hivemind proxy/deprecation is testable
with mock Dewey MCP responses. Existing Swarm test
suite validates non-Hivemind tool regression.

**Coverage strategy**:
- **Dewey repo**: Unit tests for Ollama lifecycle state
  machine, learning CRUD, embedding generation. Target
  ≥ 90% for new packages.
- **This repo**: Scaffold drift detection (existing),
  doctor check assertions (existing patterns), setup
  install command assertions. Maintain 80% global.
- **Swarm fork**: Upstream test suite + new tests for
  Hivemind proxy/deprecation. Target: zero regressions.

## Project Structure

### Documentation (this feature)

```text
specs/021-dewey-unified-memory/
├── plan.md              # This file
├── research.md          # Phase 0: design decisions
├── data-model.md        # Phase 1: learning entity, Ollama states
├── quickstart.md        # Phase 1: verification steps
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root — this repo)

```text
# Markdown files MODIFIED (agent migration):
.opencode/command/unleash.md                        # hivemind_store → dewey_store_learning
.opencode/agents/divisor-adversary.md               # hivemind_find → dewey_semantic_search
.opencode/agents/divisor-architect.md               # hivemind_find → dewey_semantic_search
.opencode/agents/divisor-guard.md                   # hivemind_find → dewey_semantic_search
.opencode/agents/divisor-sre.md                     # hivemind_find → dewey_semantic_search
.opencode/agents/divisor-testing.md                 # hivemind_find → dewey_semantic_search
.opencode/agents/cobalt-crush-dev.md                # Hivemind references → Dewey
AGENTS.md                                           # Unified memory documentation

# Scaffold asset copies (synced from live files):
internal/scaffold/assets/opencode/command/unleash.md
internal/scaffold/assets/opencode/agents/divisor-adversary.md
internal/scaffold/assets/opencode/agents/divisor-architect.md
internal/scaffold/assets/opencode/agents/divisor-guard.md
internal/scaffold/assets/opencode/agents/divisor-sre.md
internal/scaffold/assets/opencode/agents/divisor-testing.md
internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md

# Go files MODIFIED (doctor/setup/scaffold):
internal/doctor/checks.go                           # Enhanced Dewey embedding check
internal/doctor/environ.go                          # Updated Swarm install hints (fork)
internal/setup/setup.go                             # Updated Swarm install command (fork)
internal/scaffold/scaffold.go                       # Updated opencode.json plugin entry
internal/scaffold/scaffold_test.go                  # Updated assertions

# Go test files MODIFIED:
internal/doctor/doctor_test.go                      # New embedding capability test
internal/setup/setup_test.go                        # Updated install command assertions
```

### Source Code (Dewey repo — documented, not implemented here)

```text
# New package:
internal/ollama/
├── ollama.go            # Ollama lifecycle management
├── ollama_test.go       # State machine tests
└── health.go            # HTTP health check client

# Modified files:
internal/index/
├── learning.go          # NEW: Learning storage + retrieval
└── learning_test.go     # NEW: Learning CRUD tests

cmd/dewey/
└── serve.go             # Modified: Ollama auto-start on startup

# MCP tool registration:
internal/mcp/
└── tools.go             # Modified: Register dewey_store_learning
```

### Source Code (Swarm fork — documented, not implemented here)

```text
# Fork: unbound-force/swarm
# Modified files (Hivemind proxy/deprecation):
src/hivemind/
├── store.ts             # Modified: proxy to Dewey or deprecate
├── find.ts              # Modified: proxy to Dewey or deprecate
└── index.ts             # Modified: Dewey MCP client setup
```

**Structure Decision**: This is a cross-repo change.
This repo's changes are primarily Markdown edits (agent
migration) plus targeted Go modifications (doctor/setup/
scaffold). No new packages or directories in this repo.
The Dewey repo gets a new `internal/ollama/` package.
The Swarm fork modifies existing Hivemind source files.

## Phase 0: Research

See [research.md](research.md) for design decisions on:
1. **R1**: Ollama subprocess management — detect-then-start
   with fire-and-forget exit semantics
2. **R2**: Learning document schema — documents with
   `source_type: "learning"` in existing index
3. **R3**: Hivemind deprecation strategy — hard migration
   in this repo, soft deprecation (proxy) in Swarm fork
4. **R4**: Swarm fork mechanics — GitHub fork, scoped npm
   package, install command updates
5. **R5**: Doctor health check consolidation — enhanced
   Dewey embedding check, Ollama check preserved

## Phase 1: Design

See [data-model.md](data-model.md) for:
- Learning entity schema (fields, tags, query patterns)
- Ollama lifecycle state machine (states, transitions)
- Agent migration map (file-by-file Hivemind → Dewey)
- Swarm fork package identity
- Doctor health check updates

See [quickstart.md](quickstart.md) for:
- Verification steps organized by implementation phase
  (A: Dewey repo, B: this repo, C: Swarm fork,
  D: doctor/setup)

Contracts directory: **Skipped** — no new external API
or inter-hero artifact schema changes in this repo. The
`dewey_store_learning` MCP tool is defined in the Dewey
repo, not here. The learning entity uses Dewey's
existing document infrastructure, not a new schema in
the `schemas/` registry.

## Implementation Phasing

This spec MUST be implemented in three sequential
phases across repositories. Each phase is independently
deployable and testable.

### Implementation Phase 1: Dewey Repo (P1 — prerequisite)

**Scope**: FR-001 through FR-011 (Ollama lifecycle +
learning storage)
**Repo**: `unbound-force/dewey`
**Dependencies**: None (self-contained)
**Deliverables**:
- `internal/ollama/` package (detect, start, health
  check, state tracking)
- `dewey_store_learning` MCP tool
- Learning persistence in `graph.db`
- Ollama auto-start in `dewey serve`
- Tests for all new code

**Gate**: `dewey serve` auto-starts Ollama and
`dewey_store_learning` persists a learning that is
retrievable via `dewey_semantic_search`.

### Implementation Phase 2: This Repo (P1 — depends on Phase 1)

**Scope**: FR-012 through FR-015, FR-020, FR-021
(agent migration + doctor/setup)
**Repo**: `unbound-force/unbound-force`
**Dependencies**: Phase 1 complete (Dewey has
`dewey_store_learning` tool)
**Deliverables**:
- `/unleash` command updated (hivemind_store →
  dewey_store_learning)
- 5 Divisor agents updated (hivemind_find →
  dewey_semantic_search)
- AGENTS.md updated (unified memory documentation)
- Scaffold assets synced
- Doctor embedding capability check enhanced
- All tests passing

**Gate**: `grep -r "hivemind_store\|hivemind_find"
.opencode/` returns zero matches. `make check` passes.

### Implementation Phase 3: Swarm Fork (P2 — depends on Phase 2)

**Scope**: FR-016 through FR-019 (fork + setup updates)
**Repo**: `unbound-force/swarm` (new fork)
**Dependencies**: Phase 2 complete (agent files
reference Dewey, not Hivemind)
**Deliverables**:
- Forked repository at `unbound-force/swarm`
- Hivemind tools proxy through Dewey (or deprecated)
- `uf setup` installs forked plugin
- `uf doctor` install hints reference fork
- All existing Swarm tools functional (zero regressions)

**Gate**: `swarm doctor` passes. All Swarm MCP tools
functional. `uf setup` installs from fork.

## Coverage Strategy

### This Repo

- **Scaffold drift detection**: Existing
  `TestEmbeddedAssets_MatchSource` ensures modified
  `.opencode/` files match their scaffold asset copies.
- **Regression tests**: Verify no `hivemind_store` or
  `hivemind_find` references remain in scaffold assets
  (new test, pattern from
  `TestScaffoldOutput_NoGraphthulhuReferences`).
- **Doctor tests**: Existing `checkDewey` test patterns
  extended for embedding capability check.
- **Setup tests**: Existing install command assertions
  updated for forked package name.
- **No new e2e tests**: Agent file content verified by
  drift detection. Dewey integration tested in Dewey
  repo.

Coverage target: Maintain existing 80% global threshold.
New Go code (embedding capability check, fork install
commands) must have ≥ 90% coverage.

### Dewey Repo (documented, not enforced here)

- Unit tests for Ollama lifecycle state machine
- Unit tests for learning CRUD operations
- Integration test: store learning → search → verify
- Target ≥ 90% for new packages

### Swarm Fork (documented, not enforced here)

- Upstream test suite must pass (zero regressions)
- New tests for Hivemind proxy/deprecation behavior
- Target: upstream coverage maintained

## Constitution Re-Check (Post-Design)

*Re-evaluated after Phase 0 research and Phase 1
design.*

### I. Autonomous Collaboration — PASS (confirmed)

The research decision to use Dewey's existing document
index for learnings (R2) means no new inter-hero
communication protocol is introduced. Learnings are
self-describing documents with provenance metadata.
The Ollama lifecycle management (R1) is internal to
Dewey — no hero needs to know about it.

### II. Composability First — PASS (confirmed)

The research decision to leave Ollama running on exit
(R1) explicitly preserves composability — Dewey does
not assume exclusive ownership of Ollama. The Swarm
fork proxy approach (R3) ensures existing workflows
continue working during migration. The doctor check
consolidation (R5) keeps the standalone Ollama check
for independent debugging.

### III. Observable Quality — PASS (confirmed)

The learning schema (R2) includes full provenance
metadata (`source_type`, `source_id`, `created_at`,
tags). The `dewey_store_learning` tool returns
structured confirmation. The doctor embedding
capability check produces machine-parseable output.

### IV. Testability — PASS (confirmed)

All Dewey repo changes use dependency injection (HTTP
client, subprocess launcher, database). All this-repo
changes use existing injectable patterns (`LookPath`,
`ExecCmd`, `ReadFile`). The Swarm fork uses mock MCP
responses. No external services required for any test.

## Complexity Tracking

No constitution violations to justify. All changes
align with the four principles. The cross-repo nature
adds coordination complexity but does not violate any
principle — each repo's changes are independently
testable and deployable.

| Concern | Mitigation |
|---------|-----------|
| Cross-repo coordination | Sequential phasing with gates |
| Hivemind data loss | Documented in spec assumptions; volume is small |
| Swarm fork divergence | Fork is project's responsibility; upstream cherry-picks as needed |
| Ollama race condition | OS port binding provides natural mutual exclusion (R1) |
<!-- scaffolded by uf vdev -->
