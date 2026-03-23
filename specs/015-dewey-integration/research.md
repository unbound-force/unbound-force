# Research: Dewey Integration

**Date**: 2026-03-22
**Branch**: `015-dewey-integration`

## R1: MCP Configuration Format

**Decision**: Replace the graphthulhu MCP server entry
in `opencode.json` with a Dewey entry. Use the same
configuration structure.

**Rationale**: The MCP configuration is a simple JSON
object mapping server names to commands. Replacing
`knowledge-graph` with `dewey` requires changing the
server name, command, and arguments. The structure is
identical.

Before:
```json
{
  "mcp": {
    "knowledge-graph": {
      "type": "local",
      "command": ["graphthulhu", "serve", "--backend", "obsidian", "--vault", ".", "--include-hidden"],
      "enabled": true
    }
  }
}
```

After:
```json
{
  "mcp": {
    "dewey": {
      "type": "local",
      "command": ["dewey", "serve", "--vault", "."],
      "enabled": true
    }
  }
}
```

**Alternatives considered**:
- Keep both graphthulhu and Dewey: Rejected per FR-014.
  Dewey is a complete replacement.
- Conditional config (use Dewey if available, else
  graphthulhu): Rejected because it adds complexity to
  the scaffold template and the agent fallback pattern
  already handles Dewey unavailability.

## R2: Agent File Dewey Tool Pattern

**Decision**: Each hero agent file gets a "Knowledge
Retrieval" section with the 3-tier pattern:

```markdown
## Knowledge Retrieval

### When Dewey is available (full)
Use `dewey_semantic_search` for conceptual queries:
- [role-specific examples]

Use `dewey_search` for keyword queries.
Use `dewey_traverse` to follow document relationships.

### When Dewey is available without embedding model
Use `dewey_search` and `dewey_traverse` for structured
queries. Semantic search is unavailable.

### When Dewey is unavailable
Fall back to direct file reads using the Read tool.
Use Grep for keyword search across the codebase.
Reference convention packs for standards.
```

**Rationale**: The 3-tier pattern maps directly to
the spec's FR-007 and the Dewey architectural spec's
graceful degradation design (Spec 014, Research R9).
Each tier provides a functional experience with
progressively richer context.

**Alternatives considered**:
- Single-tier (Dewey or file reads): Rejected because
  the "graph-only" tier (Dewey without embedding model)
  is a common scenario during initial setup.
- Per-tool fallback (check each tool individually):
  Rejected because it fragments the agent instructions.
  Tier-based is simpler for LLMs to follow.

## R3: Role-Specific Dewey Queries

**Decision**: Each hero gets specific query examples
tailored to their role:

| Hero | Semantic Search Examples |
|------|-------------------------|
| Muti-Mind (PO) | "authentication issues across repos", "past acceptance criteria for similar features", "backlog patterns for this domain" |
| Cobalt-Crush (Dev) | "how does cobra.Command work?", "patterns for MCP tool registration", "similar implementations in other repos" |
| Gaze (Tester) | "test quality patterns in Go projects", "common CRAP score issues", "quality baselines from other repos" |
| Divisor (Reviewer) | "recurring review findings across the org", "convention violations in similar code", "architectural patterns from specs" |
| Mx F (Manager) | "velocity trends across repos", "retrospective outcomes for similar features", "coaching patterns that improved quality" |

**Rationale**: Role-specific examples help the LLM
understand what queries to make. Generic instructions
("use semantic search") are less actionable than
specific examples.

## R4: Doctor Health Check Design

**Decision**: Add a "Dewey" check group to `uf doctor`
with 3 checks:

1. **Dewey binary**: `exec.LookPath("dewey")` -- is it
   installed?
2. **Embedding model**: Check Ollama for the configured
   model -- is it pulled?
3. **Dewey workspace**: Check for `.dewey/` directory --
   is the project initialized?

Each check reports pass/fail with a fix hint.

**Rationale**: These are the 3 components a developer
needs for full Dewey functionality. Checking all three
tells the developer exactly what's missing and how to
fix it.

**Alternatives considered**:
- Single "Dewey available" check: Rejected because it
  doesn't distinguish between "not installed" and
  "installed but not initialized." The fix hints are
  different for each.
- Check Dewey index health (are documents indexed?):
  Deferred to a future enhancement. Index health
  requires the server to be running, which is complex
  for a doctor check.

## R5: Scaffold Asset Inventory

**Decision**: The following scaffold assets need tool
reference updates from `knowledge-graph_*` to `dewey_*`:

**Agent files** (check each for `knowledge-graph_` or
graphthulhu references):
- `muti-mind-po.md` -- uses knowledge graph for backlog
- `cobalt-crush-dev.md` -- may reference graphthulhu
- `gaze-reporter.md` -- likely no KG references
- `divisor-*.md` (5 files) -- may reference KG for
  cross-spec analysis
- `mx-f-coach.md` -- may reference KG for metrics
- `constitution-check.md` -- may reference KG for
  spec search
- `reviewer-*.md` (5 legacy files) -- may reference KG

**Configuration files**:
- `opencode.json` (scaffold template)

**The live copies** (`.opencode/agents/*.md`) mirror
the scaffold assets and must be updated in tandem.

**Rationale**: Tool-owned scaffold assets and their
live copies must stay in sync. The scaffold engine
deploys assets on `uf init`; the live copies in the
meta repo are the canonical versions for this repo.

## R6: Embedding Model Name for Doctor/Setup

**Decision**: Use the enterprise-grade embedding model
name in all doctor checks and setup steps. The model
name is configured in the Dewey workspace, but doctor
and setup need a default to check/pull.

**Rationale**: Doctor checks need a specific model name
to verify. The enterprise-grade model was chosen for
its Apache 2.0 license, permissible training data, and
small size (63 MB). This is the same model Dewey uses
as its default.

**Alternatives considered**:
- Check for any Ollama model: Rejected because it
  doesn't verify the correct model is installed.
- Read from Dewey config: Deferred -- this would
  require Dewey to be initialized first, creating a
  chicken-and-egg problem for the doctor check.
