# Implementation Plan: Hivemind-to-Dewey Memory Migration

**Branch**: `022-hivemind-dewey-migration` | **Date**: 2026-04-03 | **Spec**: `specs/022-hivemind-dewey-migration/spec.md`
**Input**: Feature specification from `specs/022-hivemind-dewey-migration/spec.md`

## Summary

Migrate the autonomous pipeline (`/unleash`) and all five
Divisor review agents from Hivemind to Dewey for learning
storage and retrieval. This is a Markdown-only migration
— replacing `hivemind_store` with `dewey_store_learning`
in the retrospective step, and `hivemind_find` with
`dewey_semantic_search` in the Prior Learnings step of
each Divisor agent. Scaffold asset copies must be
synchronized, documentation updated to reflect Dewey as
the unified memory layer, and a regression test added to
prevent Hivemind tool references from reappearing.

## Technical Context

**Language/Version**: Go 1.24+ (scaffold engine, regression tests), Markdown (agent files, command files, documentation)
**Primary Dependencies**: `embed.FS` (scaffold engine), `testing` (regression tests), Dewey MCP tools (`dewey_store_learning`, `dewey_semantic_search`)
**Storage**: N/A (Markdown file edits; Dewey persists to `graph.db` — out of scope for this repo)
**Testing**: `go test -race -count=1 ./...` (standard library `testing` package)
**Target Platform**: macOS, Linux (CLI tool)
**Project Type**: CLI / meta-repository (specifications, governance, agent definitions)
**Performance Goals**: N/A (no runtime performance changes — this is a text migration)
**Constraints**: All scaffold drift detection tests must pass. Zero `hivemind_store` or `hivemind_find` tool references in active agent/command files after migration.
**Scale/Scope**: 7 canonical files modified, 6 scaffold asset copies synchronized, 1 regression test added, 2 documentation files updated

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Pre-Design Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | Dewey is an artifact-based system — learnings are stored as persistent artifacts retrievable without synchronous interaction. The migration preserves the asynchronous, artifact-based communication pattern. |
| II. Composability First | PASS | All Dewey operations include graceful degradation (FR-003, FR-004). When Dewey is unavailable, agents proceed without error. No hero requires Dewey as a hard prerequisite. |
| III. Observable Quality | PASS | Dewey's `dewey_store_learning` persists learnings with metadata (tags, timestamps). Learnings are retrievable via semantic search, maintaining provenance. |
| IV. Testability | PASS | The migration is testable via: (a) text search for stale tool references (SC-001), (b) scaffold drift detection tests (SC-005), (c) a new `TestScaffoldOutput_NoHivemindReferences` regression test. All tests run in isolation without external services. |

**Gate result**: PASS — no violations. Proceed to Phase 0.

### Post-Design Check

| Principle | Status | Evidence |
|-----------|--------|----------|
| I. Autonomous Collaboration | PASS | Unchanged from pre-design. Dewey stores learnings as self-describing artifacts with tags and metadata. |
| II. Composability First | PASS | Graceful degradation preserved in all modified files. Three tiers: Full Dewey, Dewey without learning storage (dewey#25 not landed), No Dewey. |
| III. Observable Quality | PASS | Regression test provides automated, reproducible evidence that Hivemind references are eliminated. Drift detection tests verify scaffold synchronization. |
| IV. Testability | PASS | Coverage strategy defined: regression test for stale references, existing drift detection for scaffold sync. All tests isolated — no external services needed. |

**Gate result**: PASS — no violations.

## Project Structure

### Documentation (this feature)

```text
specs/022-hivemind-dewey-migration/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── tasks.md             # Phase 2 output (/speckit.tasks)
└── checklists/
    └── requirements.md  # Spec quality checklist
```

### Source Code (repository root)

```text
# Files modified by this feature (Markdown migration + Go test)

.opencode/
├── command/
│   └── unleash.md                    # FR-001: hivemind_store → dewey_store_learning
└── agents/
    ├── divisor-adversary.md          # FR-002: hivemind_find → dewey_semantic_search
    ├── divisor-architect.md          # FR-002: hivemind_find → dewey_semantic_search
    ├── divisor-guard.md              # FR-002: hivemind_find → dewey_semantic_search
    ├── divisor-sre.md                # FR-002: hivemind_find → dewey_semantic_search
    └── divisor-testing.md            # FR-002: hivemind_find → dewey_semantic_search

internal/scaffold/
├── assets/
│   ├── opencode/command/
│   │   └── unleash.md                # FR-006: scaffold copy sync
│   └── opencode/agents/
│       ├── divisor-adversary.md      # FR-006: scaffold copy sync
│       ├── divisor-architect.md      # FR-006: scaffold copy sync
│       ├── divisor-guard.md          # FR-006: scaffold copy sync
│       ├── divisor-sre.md            # FR-006: scaffold copy sync
│       └── divisor-testing.md        # FR-006: scaffold copy sync
└── scaffold_test.go                  # FR-007: NoHivemindReferences test

AGENTS.md                             # FR-005: documentation update
```

**Structure Decision**: No new directories or packages. This is a
text migration across existing files. The scaffold engine and test
infrastructure already exist — we add one regression test function
to the existing `scaffold_test.go` file, following the pattern
established by `TestScaffoldOutput_NoGraphthulhuReferences`.

## Coverage Strategy

Per Constitution Principle IV (Testability), the coverage
strategy for this migration:

| Layer | What | How |
|-------|------|-----|
| Regression | No `hivemind_store`/`hivemind_find` in scaffold assets | `TestScaffoldOutput_NoHivemindReferences` (new) |
| Drift | Scaffold assets match canonical sources | `TestEmbeddedAssets_MatchSource` (existing) |
| Smoke | `go test -race -count=1 ./...` passes | CI pipeline |
| Manual | End-to-end `/unleash` with Dewey | Quickstart verification steps |

No new unit tests for business logic are needed — this
migration modifies Markdown instruction files, not Go
source code. The regression test is the primary automated
gate.

## Complexity Tracking

No constitution violations to justify. All changes are
straightforward text replacements within existing files.
<!-- scaffolded by uf vdev -->
