# Data Model: Hivemind-to-Dewey Memory Migration

**Branch**: `022-hivemind-dewey-migration`
**Date**: 2026-04-03

## Overview

This migration does not introduce new data entities. It
changes the tool references in agent instruction files
to point at Dewey instead of Hivemind. The data model
documents the tool mapping, the file migration map, and
the text replacement patterns.

## Tool Migration Map

Mapping from Hivemind tools to Dewey equivalents, per
Spec 021 data-model.md.

| Hivemind Tool | Dewey Replacement | Status | Notes |
|--------------|-------------------|--------|-------|
| `hivemind_store` | `dewey_store_learning` | Pending (dewey#25) | New Dewey MCP tool for learning persistence |
| `hivemind_find` | `dewey_semantic_search` | Available | Existing Dewey tool; already used for knowledge retrieval |

Only the two tools above are in scope for this spec.
Other Hivemind tools (`hivemind_validate`,
`hivemind_remove`, `hivemind_sync`, `hivemind_stats`,
`hivemind_index`) are not referenced in the files being
migrated.

## File Migration Map

### Canonical Files (source of truth)

| File | Current Reference | Replacement | FR | US |
|------|------------------|-------------|-----|-----|
| `.opencode/command/unleash.md` | `hivemind_store` | `dewey_store_learning` | FR-001 | US-1 |
| `.opencode/agents/divisor-adversary.md` | `hivemind_find` | `dewey_semantic_search` | FR-002 | US-2 |
| `.opencode/agents/divisor-architect.md` | `hivemind_find` | `dewey_semantic_search` | FR-002 | US-2 |
| `.opencode/agents/divisor-guard.md` | `hivemind_find` | `dewey_semantic_search` | FR-002 | US-2 |
| `.opencode/agents/divisor-sre.md` | `hivemind_find` | `dewey_semantic_search` | FR-002 | US-2 |
| `.opencode/agents/divisor-testing.md` | `hivemind_find` | `dewey_semantic_search` | FR-002 | US-2 |

### Scaffold Asset Copies (must sync with canonical)

| Scaffold Asset | Canonical Source |
|---------------|-----------------|
| `internal/scaffold/assets/opencode/command/unleash.md` | `.opencode/command/unleash.md` |
| `internal/scaffold/assets/opencode/agents/divisor-adversary.md` | `.opencode/agents/divisor-adversary.md` |
| `internal/scaffold/assets/opencode/agents/divisor-architect.md` | `.opencode/agents/divisor-architect.md` |
| `internal/scaffold/assets/opencode/agents/divisor-guard.md` | `.opencode/agents/divisor-guard.md` |
| `internal/scaffold/assets/opencode/agents/divisor-sre.md` | `.opencode/agents/divisor-sre.md` |
| `internal/scaffold/assets/opencode/agents/divisor-testing.md` | `.opencode/agents/divisor-testing.md` |

### Documentation Files

| File | Change | FR |
|------|--------|-----|
| `AGENTS.md` | Update "Embedding Model Alignment" section, Spec 020 Recent Changes entry, add Spec 022 Recent Changes entry | FR-005 |
| `internal/setup/setup.go` | Update comments referencing Hivemind | FR-005 |

## Text Replacement Patterns

### `/unleash` Retrospective Step (FR-001)

The retrospective step (Step 7, lines ~484-519 of
`unleash.md`) contains these Hivemind references that
must be replaced:

| Line Context | Old Text | New Text |
|-------------|----------|----------|
| Learning format guidance | "Hivemind's semantic search works best on narrative text" | "Dewey's semantic search works best on narrative text" |
| Tool availability check | `hivemind_store` tool exists | `dewey_store_learning` tool exists |
| Storage call | `hivemind_store` with tags | `dewey_store_learning` with tags |
| Unavailable message | "Hivemind not available" | "Dewey not available" |
| Unavailable hint | "Install the Swarm plugin for semantic memory" | "Install Dewey for semantic memory" |
| Section heading check | "If Hivemind is available" | "If Dewey is available" |
| Section heading check | "If Hivemind is NOT available" | "If Dewey is NOT available" |

### Divisor Agent Prior Learnings Step (FR-002)

Each of the five Divisor agents has an identical "Step 0:
Prior Learnings" section with these Hivemind references:

| Line Context | Old Text | New Text |
|-------------|----------|----------|
| Tool availability check | "If Hivemind MCP tools are available (`hivemind_find`)" | "If Dewey MCP tools are available (`dewey_semantic_search`)" |
| Query call | `hivemind_find({ query: "..." })` | `dewey_semantic_search({ query: "..." })` |
| Unavailable fallback | "If Hivemind is not available" | "If Dewey is not available" |

### AGENTS.md Documentation (FR-005)

| Section | Old Text | New Text |
|---------|----------|----------|
| Embedding Model Alignment | "To ensure Swarm's Hivemind uses the same model" | "To ensure all tools use the same model" |
| Spec 020 Recent Changes | "Dewey complements Hivemind (Spec 019), not replaces it" | "Dewey is the unified memory layer (superseded by Spec 022)" |

### setup.go Comments (FR-005)

| Line | Old Comment | New Comment |
|------|------------|-------------|
| ~112 | "Setting these env vars aligns Swarm's Hivemind with Dewey's embedding model" | "Setting these env vars aligns all tools with Dewey's embedding model" |
| ~128 | "Set Ollama env vars so Swarm's Hivemind uses the same" | "Set Ollama env vars so all embedding consumers use the same" |

## Regression Test Specification

### `TestScaffoldOutput_NoHivemindReferences`

**Location**: `internal/scaffold/scaffold_test.go`

**Pattern**: Follows
`TestScaffoldOutput_NoGraphthulhuReferences` exactly.

**Stale patterns** (must NOT appear in scaffolded output):
- `hivemind_store`
- `hivemind_find`
- `hivemind_validate`
- `hivemind_remove`
- `hivemind_get`

**Rationale for broad pattern list**: While only
`hivemind_store` and `hivemind_find` are currently
referenced in the migrated files, checking all Hivemind
tool names prevents accidental future introduction of
any Hivemind tool reference in scaffold assets. The
cost is negligible (3 extra string comparisons per file)
and the protection is comprehensive. `hivemind_sync`,
`hivemind_stats`, and `hivemind_index` are excluded
because they are Swarm plugin infrastructure tools, not
agent-facing learning tools — their presence in scaffold
assets would be a different concern.

**Implementation**:
1. Scaffold to `t.TempDir()` via `Run(Options{...})`
2. Walk all generated files
3. For each file, check for stale patterns
4. Fail with descriptive error if any match found

**Note**: The test checks scaffolded output (the files
deployed by `uf init`), not the repo root. This means
AGENTS.md is not checked by this test (it is not a
scaffold asset). AGENTS.md is checked by SC-004 via
manual verification and the quickstart steps.

## Cobalt-Crush Agent Reference

The `cobalt-crush-dev.md` agent file contains one
Hivemind prose reference on line 190:

> "complementing Hivemind's session-specific learnings"

This reference is in the Knowledge Retrieval section's
description of Dewey's role. Per FR-005, this should be
updated to reflect Dewey as the unified memory layer.
However, `cobalt-crush-dev.md` does not reference
`hivemind_store` or `hivemind_find` as tool calls — the
reference is purely descriptive prose. The update is
included in the documentation phase (FR-005) rather than
the tool migration phase (FR-001/FR-002).

## AGENTS.md Recent Changes Entry (Draft)

Template for T017 — add to the top of the Recent Changes
section in `AGENTS.md`:

> - 022-hivemind-dewey-migration: Migrated `/unleash`
>   retrospective and all five Divisor review agents from
>   Hivemind to Dewey for learning storage and retrieval.
>   Replaced `hivemind_store` with `dewey_store_learning`
>   in the autonomous pipeline retrospective step (FR-001).
>   Replaced `hivemind_find` with `dewey_semantic_search`
>   in the Prior Learnings step of all five `divisor-*.md`
>   agents (FR-002). Updated AGENTS.md and setup.go to
>   describe Dewey as the unified memory layer, superseding
>   the "Dewey complements Hivemind" framing from Spec 020
>   (FR-005). Synchronized 7 scaffold asset copies (FR-006).
>   Added `TestScaffoldOutput_NoHivemindReferences`
>   regression test (FR-007). All 5 user stories and 29
>   tasks completed.
