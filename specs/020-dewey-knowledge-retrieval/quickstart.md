# Quickstart: Dewey Knowledge Retrieval

**Branch**: `020-dewey-knowledge-retrieval`
**Date**: 2026-04-01

## Prerequisites

- Repository scaffolded with `uf init`
- Dewey installed (`brew install unbound-force/tap/dewey`
  or via `uf setup`)
- Dewey workspace initialized (`dewey init` in repo root)
- Embedding model pulled (`ollama pull granite-embedding:30m`)

## Verification Steps

### 1. Verify AGENTS.md Knowledge Retrieval Section

**Test**: Open AGENTS.md and search for "Knowledge Retrieval".

**Expected**: A section titled "Knowledge Retrieval" exists
between "Coding Conventions" and "Testing Conventions" that:
- Instructs agents to prefer Dewey MCP tools over grep/glob
- Lists the tool selection matrix (semantic search, keyword
  search, get_page, find_connections)
- Specifies when to fall back to grep/glob/read
- Includes the 3-tier graceful degradation pattern

### 2. Verify Cobalt-Crush Knowledge Step

**Test**: Read `.opencode/agents/cobalt-crush-dev.md` and
look for the Knowledge Retrieval section.

**Expected**: The existing Knowledge Retrieval section is
enhanced with:
- A "Step 0: Knowledge Retrieval" instruction that fires
  before code exploration
- Queries for prior learnings about target files
- Queries for related specs governing the feature
- Queries for architectural patterns from conventions
- 3-tier graceful degradation

### 3. Verify Speckit Command Integration

**Test**: Read the three Speckit command files:
- `.opencode/command/speckit.specify.md`
- `.opencode/command/speckit.plan.md`
- `.opencode/command/speckit.tasks.md`

**Expected**: Each command includes a Dewey query step:
- specify: queries for existing specs with similar topics
- plan: queries for prior research decisions
- tasks: queries for implementation patterns

Each query step includes graceful degradation (skip if
Dewey unavailable).

### 4. Verify All Hero Agent Knowledge Retrieval

**Test**: Read each hero agent file:
- `.opencode/agents/muti-mind-po.md`
- `.opencode/agents/mx-f-coach.md`
- `.opencode/agents/gaze-reporter.md`

**Expected**: Each agent's existing Knowledge Retrieval
section is enhanced with role-appropriate query examples
and the "prefer Dewey" behavioral instruction. The 3-tier
graceful degradation pattern is present.

### 5. Verify Heroes Skill Update

**Test**: Read `.opencode/skill/unbound-force-heroes/SKILL.md`.

**Expected**: The skill mentions Dewey as the knowledge
retrieval layer in the hero lifecycle documentation.
Swarm coordinators are instructed to query Dewey for
relevant context before each stage.

### 6. Verify Scaffold Asset Sync

**Test**: Run `make test` or `go test -race -count=1 ./...`

**Expected**: All tests pass, including scaffold drift
detection tests. Every modified live file has a matching
copy in `internal/scaffold/assets/`.

### 7. Verify Graceful Degradation

**Test**: In a repo without Dewey configured, open a new
OpenCode session and ask "how does the scaffold system
work?"

**Expected**: The agent attempts Dewey queries, finds
Dewey unavailable, falls back to grep/glob/read with no
error or interruption. The agent still answers the
question using direct file access.

### 8. Verify File Count Unchanged

**Test**: Count scaffold assets.

```bash
find internal/scaffold/assets -type f | wc -l
```

**Expected**: Same count as before this spec (no files
added or removed).

## Smoke Test Sequence

For a quick end-to-end verification:

1. `grep -c "Knowledge Retrieval" AGENTS.md` → should
   return at least 1
2. `grep -c "dewey_semantic_search" .opencode/agents/cobalt-crush-dev.md`
   → should return at least 1
3. `grep -c "dewey_" .opencode/command/speckit.specify.md`
   → should return at least 1
4. `make test` → all tests pass
5. `diff .opencode/agents/cobalt-crush-dev.md internal/scaffold/assets/opencode/agents/cobalt-crush-dev.md`
   → no differences
