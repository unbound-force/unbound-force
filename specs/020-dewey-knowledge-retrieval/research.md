# Research: Dewey Knowledge Retrieval

**Branch**: `020-dewey-knowledge-retrieval`
**Date**: 2026-04-01

## Research Question 1: How to Structure the AGENTS.md Knowledge Retrieval Section

### Decision

Add a top-level "Knowledge Retrieval" section to AGENTS.md
(after "Coding Conventions" and before "Testing Conventions")
that establishes a project-wide convention: agents SHOULD
prefer Dewey MCP tools over grep/glob for cross-repo context,
design decisions, and architectural patterns.

### Rationale

AGENTS.md is the primary context file injected into every
OpenCode agent session. A behavioral instruction here is the
highest-leverage change — it affects every agent, every
command, every session. The section should be:

1. **Concise** — agents have limited context windows. The
   section should be 20-30 lines, not a full tutorial.
2. **Prescriptive** — tell agents *when* to use which tool,
   not just *that* Dewey exists.
3. **Graceful** — include the 3-tier degradation pattern
   so agents work without Dewey.

### Alternatives Considered

- **Per-agent only** (no AGENTS.md section): Rejected because
  new agents wouldn't inherit the convention. AGENTS.md is
  the canonical source of project-wide conventions.
- **Detailed tutorial in AGENTS.md**: Rejected because it
  would bloat the context window. Keep AGENTS.md concise;
  detailed per-role guidance lives in each agent file.
- **Separate knowledge-retrieval.md file**: Rejected because
  it adds a file agents must read. AGENTS.md is already
  loaded; adding content there is zero-cost.

## Research Question 2: Dewey Tool Selection Matrix

### Decision

Map query types to Dewey tools based on the query's nature:

| Query Type | Dewey Tool | When to Use | Example |
|------------|-----------|-------------|---------|
| Conceptual / "how does X work?" | `dewey_semantic_search` | Understanding patterns, finding similar implementations | "how does the scaffold system work?" |
| Keyword / exact term | `dewey_search` | Finding specific references, known terms | "scaffold.go", "FR-001" |
| Specific page | `dewey_get_page` | Reading a known spec or document | "specs/014-dewey-architecture/spec" |
| Relationship discovery | `dewey_find_connections` | Understanding how concepts relate | "how are Gaze and Mx F connected?" |
| Similar documents | `dewey_similar` | Finding related specs or patterns | "find specs similar to this one" |
| Tag-based discovery | `dewey_find_by_tag` | Finding all items with a tag | "all pages tagged #decision" |
| Property queries | `dewey_query_properties` | Finding items by metadata | "all specs with status: draft" |
| Filtered semantic | `dewey_semantic_search_filtered` | Semantic search within a source type | "authentication patterns from GitHub sources" |

### Rationale

Agents need a clear decision tree, not a list of tools.
The matrix maps the *intent* of the query to the *tool*
that best serves it. This prevents agents from defaulting
to `dewey_search` for everything (which misses semantic
matches) or `dewey_semantic_search` for everything (which
is slower for exact lookups).

### Alternatives Considered

- **Single tool recommendation** (always use semantic search):
  Rejected because keyword search is faster and more precise
  for exact terms. Semantic search adds latency and may
  return false positives for specific lookups.
- **No matrix, just examples**: Rejected because agents
  perform better with structured decision criteria than
  with examples alone.

## Research Question 3: Cobalt-Crush Knowledge Retrieval Step Integration

### Decision

Add a "Step 0: Knowledge Retrieval" to Cobalt-Crush's
workflow that fires *before* reading source documents.
The step queries Dewey for:

1. **Prior learnings** about the target files
   (`dewey_semantic_search` for file-specific context)
2. **Related specs** that govern the feature being
   implemented (`dewey_search` for spec references)
3. **Architectural patterns** from conventions
   (`dewey_find_by_tag` for convention-tagged content)

This mirrors the Divisor agents' "Step 0: Prior Learnings"
pattern (from Spec 019) but uses Dewey instead of Hivemind.
The two are complementary: Hivemind stores session-specific
learnings; Dewey provides cross-repo architectural context.

### Rationale

Cobalt-Crush is the primary implementation agent. Querying
Dewey before coding grounds implementations in project
history and conventions. The "Step 0" pattern is already
established by the Divisor agents (Spec 019), so this
follows a proven pattern.

### Alternatives Considered

- **Integrate into Source Documents section**: Rejected
  because Source Documents is about reading specific files.
  Knowledge Retrieval is about querying for *unknown*
  relevant context — a fundamentally different operation.
- **Add after code exploration**: Rejected because the
  value of Dewey context is highest *before* the agent
  forms its implementation plan. Late context leads to
  rework.
- **Replace Hivemind with Dewey**: Rejected per spec
  assumption — Hivemind stores session-specific learnings;
  Dewey provides cross-repo architectural context. They
  serve different purposes.

## Research Question 4: Speckit Command Dewey Integration

### Decision

Add a Dewey query step to three Speckit commands:

1. **`/speckit.specify`**: Before generating the spec,
   query Dewey for existing specs with similar topics
   (`dewey_semantic_search`). Reference discovered specs
   in the Dependencies section. This prevents duplicate
   or conflicting specifications.

2. **`/speckit.plan`**: During Phase 0 research, query
   Dewey for prior research decisions in related specs
   (`dewey_search` for research.md files). This grounds
   research in existing project knowledge.

3. **`/speckit.tasks`**: Before generating tasks, query
   Dewey for implementation patterns from completed specs
   (`dewey_semantic_search_filtered` for completed specs).
   This helps generate more realistic task breakdowns.

### Rationale

The Speckit pipeline generates foundational artifacts.
Grounding these in project history reduces rework and
inconsistency. Each command gets a targeted query type
appropriate to its function.

### Alternatives Considered

- **Add Dewey to all 8 Speckit commands**: Rejected per
  YAGNI. Only specify, plan, and tasks benefit from
  cross-repo context. Commands like clarify, analyze,
  and checklist operate on existing artifacts.
- **Add Dewey only to specify**: Rejected because plan
  and tasks also benefit from historical context (prior
  research decisions, implementation patterns).

## Research Question 5: Graceful Degradation Pattern

### Decision

Use the established 3-tier pattern from Spec 015:

```
Tier 3 (Full Dewey) — semantic + structured search
Tier 2 (Graph-only, no embedding model) — structured only
Tier 1 (No Dewey) — direct file access
```

Each Knowledge Retrieval section MUST include all three
tiers. The agent detects availability by attempting the
query — if it fails, it falls back to the next tier.
No explicit availability check is needed.

### Rationale

This pattern is already established across all hero agents
(Spec 015) and the Divisor agents (Spec 019). Consistency
is more important than optimization. The "attempt and
fallback" approach is simpler than explicit health checks
and handles partial availability (e.g., Dewey running but
embedding model not loaded).

### Alternatives Considered

- **Explicit health check before queries**: Rejected
  because it adds complexity and doesn't handle partial
  availability well.
- **2-tier pattern** (Dewey or no Dewey): Rejected because
  the graph-only tier is a real scenario (Dewey running
  without an embedding model configured).
