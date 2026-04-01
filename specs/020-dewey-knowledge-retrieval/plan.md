# Implementation Plan: Dewey Knowledge Retrieval

**Branch**: `020-dewey-knowledge-retrieval` | **Date**: 2026-04-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/020-dewey-knowledge-retrieval/spec.md`

## Summary

Add behavioral instructions to AGENTS.md, all hero agent files,
Speckit command files, and the heroes skill so that AI agents
prefer Dewey MCP tools for cross-repo context, design decisions,
and architectural patterns. This is a Markdown-only change with
scaffold asset syncing — no new Go production code, no file count
changes.

## Technical Context

**Language/Version**: Markdown (no Go code changes)
**Primary Dependencies**: Dewey MCP server (optional at runtime)
**Storage**: N/A (Markdown files deployed to target directory)
**Testing**: Scaffold drift detection tests (existing `go test`)
**Target Platform**: OpenCode agent runtime (any OS)
**Project Type**: Meta repository (specifications, governance,
agent behavioral instructions)
**Performance Goals**: N/A (documentation changes only)
**Constraints**: Scaffold file count MUST remain unchanged
(no files added or removed). All modified files MUST be synced
to `internal/scaffold/assets/`.
**Scale/Scope**: 10+ Markdown files modified, 0 Go files changed

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after
Phase 1 design.*

### I. Autonomous Collaboration — PASS

Heroes communicate through well-defined artifacts. This spec
adds behavioral instructions that direct agents to query Dewey
for context *before* performing their primary function. Dewey
queries are asynchronous artifact reads — no runtime coupling
between heroes. Each hero's Knowledge Retrieval section is
self-contained; no hero depends on another hero's Dewey query.

### II. Composability First — PASS

**Key principle for this spec.** Every Knowledge Retrieval
instruction includes a 3-tier graceful degradation pattern:
- Tier 3 (Full Dewey): semantic + structured search
- Tier 2 (Graph-only): structured search only
- Tier 1 (No Dewey): direct file access via Read/Grep

This ensures every agent works standalone without Dewey.
Dewey is an optional enhancement, never a hard dependency.
The "prefer Dewey" instruction is a SHOULD (soft preference),
not a MUST (hard requirement). Agents retain discretion to
use grep/glob/read when more appropriate.

### III. Observable Quality — PASS

No new artifacts are produced by this spec. The behavioral
instructions direct agents to *consume* Dewey's existing
machine-parseable output (search results with provenance
metadata, similarity scores, source type indicators). Dewey
already satisfies Observable Quality through its structured
JSON responses.

### IV. Testability — PASS

No new Go code is written. The only "test" is the existing
scaffold drift detection test (`TestScaffoldOutput_*`) which
verifies that live files match their scaffold asset copies.
This test already exists and will catch any sync drift. No
new coverage strategy is needed because there is no new code.

**Coverage strategy**: N/A — Markdown-only change. Existing
scaffold drift detection tests provide the verification gate.

## Project Structure

### Documentation (this feature)

```text
specs/020-dewey-knowledge-retrieval/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (tool selection matrix)
├── quickstart.md        # Phase 1 output (verification steps)
└── checklists/          # Spec quality checklist (from /speckit.specify)
```

### Source Code (repository root)

```text
# Files MODIFIED (not created or deleted):
AGENTS.md                                          # Knowledge Retrieval convention section
.opencode/agents/cobalt-crush-dev.md               # Enhanced Knowledge Retrieval step
.opencode/agents/muti-mind-po.md                   # Enhanced Knowledge Retrieval step
.opencode/agents/mx-f-coach.md                     # Enhanced Knowledge Retrieval step
.opencode/agents/gaze-reporter.md                  # Enhanced Knowledge Retrieval step
.opencode/command/speckit.specify.md               # New Dewey query step
.opencode/command/speckit.plan.md                  # New Dewey query step
.opencode/command/speckit.tasks.md                 # New Dewey query step
.opencode/skill/unbound-force-heroes/SKILL.md      # Dewey in hero lifecycle

# Scaffold asset copies (synced from live files):
internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md
internal/scaffold/assets/opencode/agents/muti-mind-po.md
internal/scaffold/assets/opencode/agents/mx-f-coach.md
internal/scaffold/assets/opencode/agents/gaze-reporter.md
internal/scaffold/assets/opencode/command/speckit.specify.md
internal/scaffold/assets/opencode/command/speckit.plan.md
internal/scaffold/assets/opencode/command/speckit.tasks.md
internal/scaffold/assets/opencode/skill/unbound-force-heroes/SKILL.md
```

**Structure Decision**: No new directories or files. All changes
are edits to existing Markdown files plus syncing those edits to
their scaffold asset copies under `internal/scaffold/assets/`.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
<!-- scaffolded by uf vdev -->
