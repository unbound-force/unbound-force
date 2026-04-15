# Data Model: AGENTS.md Behavioral Guidance Injection

**Branch**: `030-agents-md-guidance` | **Date**: 2026-04-15
**Phase**: 1 (Design)

## Overview

This feature has no traditional data model — it does not
introduce new Go structs, JSON schemas, database tables,
or API endpoints. The "data" is Markdown text injected
into `AGENTS.md` files.

This document defines the **guidance block schema** — the
structure and content of each of the 8 standardized
blocks that `/uf-init` injects.

## Guidance Block Schema

Each guidance block has the following properties:

| Property | Type | Description |
|----------|------|-------------|
| `id` | string | Unique identifier (e.g., `core-mission`) |
| `heading` | string | Markdown heading text |
| `heading_level` | int | Heading level (2 = `##`, 3 = `###`) |
| `detection_phrases` | string[] | Phrases to search for in idempotency check |
| `placement` | string | Where to insert in AGENTS.md |
| `priority` | P1/P2/P3 | Implementation priority (from spec) |
| `content` | string | The Markdown text to inject |
| `source` | string | Canonical source file and lines |

## Block Definitions

### Block 1: Core Mission

```yaml
id: core-mission
heading: "## Core Mission"
heading_level: 2
detection_phrases:
  - "## Core Mission"
  - "Strategic Architecture"
  - "Outcome Orientation"
placement: "After ## Project Overview, before ## Behavioral Constraints"
priority: P3
source: "meta AGENTS.md lines 12-16, gaze AGENTS.md lines 12-16"
```

**Content** (generalized):

```markdown
## Core Mission

- **Strategic Architecture**: Engineers shift from manual
  coding to directing an "infinite supply of junior
  developers" (AI agents).
- **Outcome Orientation**: Focus on conveying business
  value and user intent rather than low-level technical
  sub-tasks.
- **Intent-to-Context**: Treat specs and rules as the
  medium through which human intent is manifested into
  code.
```

### Block 2: Gatekeeping Value Protection

```yaml
id: gatekeeping-value-protection
heading: "### Gatekeeping Value Protection"
heading_level: 3
detection_phrases:
  - "Gatekeeping Value Protection"
  - "MUST NOT modify values that serve as quality"
placement: "Inside ## Behavioral Constraints"
priority: P1
source: "meta AGENTS.md lines 25-38"
```

**Content** (generalized):

```markdown
### Gatekeeping Value Protection

Agents MUST NOT modify values that serve as quality or
governance gates to make an implementation pass. The
following categories are protected:

1. **Coverage thresholds and CRAP scores** — minimum
   coverage percentages, CRAP score limits, coverage
   ratchets
2. **Severity definitions and auto-fix policies** —
   CRITICAL/HIGH/MEDIUM/LOW boundaries, auto-fix
   eligibility rules
3. **Convention pack rule classifications** —
   MUST/SHOULD/MAY designations on convention pack rules
   (downgrading MUST to SHOULD is prohibited)
4. **CI flags and linter configuration** — `-race`,
   `-count=1`, `govulncheck`, `golangci-lint` rules,
   pinned action SHAs
5. **Agent temperature and tool-access settings** —
   frontmatter `temperature`, `tools.write`, `tools.edit`,
   `tools.bash` restrictions
6. **Constitution MUST rules** — any MUST rule in
   `.specify/memory/constitution.md` or hero constitutions
7. **Review iteration limits and worker concurrency** —
   max review iterations, max concurrent Swarm workers,
   retry limits
8. **Workflow gate markers** — `<!-- spec-review: passed
   -->`, task completion checkboxes used as gates, phase
   checkpoint requirements

**What to do instead**: When an implementation cannot
meet a gate, the agent MUST stop, report which gate is
blocking and why, and let the human decide whether to
adjust the gate or rework the implementation. Modifying
a gate without explicit human authorization is a
constitution violation (CRITICAL severity).
```

### Block 3: Workflow Phase Boundaries

```yaml
id: workflow-phase-boundaries
heading: "### Workflow Phase Boundaries"
heading_level: 3
detection_phrases:
  - "Workflow Phase Boundaries"
  - "MUST NOT cross workflow phase boundaries"
placement: "Inside ## Behavioral Constraints, after Gatekeeping"
priority: P1
source: "meta AGENTS.md lines 40-54"
```

**Content** (generalized):

```markdown
### Workflow Phase Boundaries

Agents MUST NOT cross workflow phase boundaries:

- **Specify/Clarify/Plan/Tasks/Analyze/Checklist** phases:
  spec artifacts ONLY (`specs/NNN-*/` directory). No
  source code, test, agent, command, or config changes.
- **Implement** phase: source code changes allowed,
  guided by spec artifacts.
- **Review** phase: findings and minor fixes only. No new
  features.

A phase boundary violation is treated as a process error.
The agent MUST stop and report the violation rather than
proceeding with out-of-phase changes.
```

### Block 4: CI Parity Gate

```yaml
id: ci-parity-gate
heading: "CI Parity Gate"
heading_level: 3  # or bold bullet
detection_phrases:
  - "CI Parity Gate"
  - "replicate the CI checks locally"
placement: "Inside ## Technical Guardrails or ## Behavioral Constraints"
priority: P1
source: "gaze AGENTS.md line 27"
```

**Content** (generalized):

```markdown
### CI Parity Gate

Before marking any implementation task complete or
declaring a PR ready, agents MUST replicate the CI checks
locally. Read `.github/workflows/` to identify the exact
commands CI runs, then execute those same commands. Any
failure is a blocking error — a task is not complete
until all CI-equivalent checks pass locally. Do not rely
on a memorized list of commands; always derive them from
the workflow files, which are the source of truth.
```

### Block 5: Review Council PR Prerequisite

```yaml
id: review-council-pr-prerequisite
heading: "### Review Council as PR Prerequisite"
heading_level: 3
detection_phrases:
  - "Review Council"
  - "PR Prerequisite"
  - "/review-council"
placement: "After behavioral constraints, before build commands"
priority: P2
source: "gaze AGENTS.md lines 38-54"
```

**Content** (generalized):

```markdown
### Review Council as PR Prerequisite

Before submitting a pull request, agents **must** run
`/review-council` and resolve all REQUEST CHANGES
findings until all reviewers return APPROVE. There must
be **minimal to no code changes** between the council's
APPROVE verdict and the PR submission — the council
reviews the final code, not a draft that changes
afterward.

Workflow:

1. Complete all implementation tasks
2. Run CI checks locally (build, test, vet)
3. Run `/review-council` — fix any findings, re-run
   until APPROVE
4. Commit, push, and submit PR immediately after council
   APPROVE
5. Do NOT make further code changes between APPROVE and
   PR submission

Exempt from council review:

- Constitution amendments (governance documents, not code)
- Documentation-only changes (README, AGENTS.md, spec
  artifacts)
- Emergency hotfixes (must be retroactively reviewed)
```

### Block 6: Website Documentation Sync Gate

```yaml
id: website-documentation-sync-gate
heading: "### Website Documentation Gate"
heading_level: 3
detection_phrases:
  - "Website Documentation"
  - "gh issue create --repo"
  - "unbound-force/website"
placement: "Near documentation validation gate or spec commit gate"
priority: P2
source: "meta AGENTS.md lines 397-418"
```

**Content** (generalized):

````markdown
### Website Documentation Gate

When a change affects user-facing behavior, hero
capabilities, CLI commands, or workflows, a GitHub issue
**MUST** be created in the `unbound-force/website`
repository to track required documentation or website
updates. The issue must be created before the
implementing PR is merged.

```bash
gh issue create --repo unbound-force/website \
  --title "docs: <brief description of what changed>" \
  --body "<what changed, why it matters, which pages
          need updating>"
```

**Exempt changes** (no website issue needed):
- Internal refactoring with no user-facing behavior
  change
- Test-only changes
- CI/CD pipeline changes
- Spec artifacts (specs are internal planning documents)

**Examples requiring a website issue**:
- New CLI command or flag added
- Hero capabilities changed (new agent, removed feature)
- Installation steps changed (`uf setup` flow)
- New convention pack added
- Breaking changes to any user-facing workflow
````

### Block 7: Spec-First Development

```yaml
id: spec-first-development
heading: "## Spec-First Development"
heading_level: 2
detection_phrases:
  - "Spec-First Development"
  - "preceded by a spec workflow"
placement: "After behavioral constraints, before build commands"
priority: P2
source: "gaze AGENTS.md lines 56-82"
```

**Content** (generalized):

```markdown
## Spec-First Development

All changes that modify production code, test code, agent
prompts, embedded assets, or CI configuration **must** be
preceded by a spec workflow. The constitution
(`.specify/memory/constitution.md`) is the highest-
authority document in this project — all work must align
with it.

Two spec workflows are available:

| Workflow | Location | Best For |
|----------|----------|----------|
| **Speckit** | `specs/NNN-name/` | Numbered feature specs with the full pipeline |
| **OpenSpec** | `openspec/changes/name/` | Targeted changes with lightweight artifacts |

**What requires a spec** (no exceptions without explicit
user override):

- New features or capabilities
- Refactoring that changes function signatures, extracts
  helpers, or moves code between packages
- Test additions or assertion strengthening across
  multiple functions
- Agent prompt changes
- CI workflow modifications
- Data model changes (new struct fields, schema updates)

**What is exempt** (may be done directly):

- Constitution amendments (governed by the constitution's
  own Governance section)
- Typo corrections, comment-only changes, single-line
  formatting fixes
- Emergency hotfixes for critical production bugs (must
  be retroactively documented)

When an agent is unsure whether a change is trivial, it
**must** ask the user rather than proceeding without a
spec. The cost of an unnecessary spec is minutes; the
cost of an unplanned change is rework, drift, and broken
CI.
```

### Block 8: Knowledge Retrieval

```yaml
id: knowledge-retrieval
heading: "## Knowledge Retrieval"
heading_level: 2
detection_phrases:
  - "## Knowledge Retrieval"
  - "dewey_semantic_search"
  - "Tool Selection Matrix"
placement: "After coding conventions, before testing conventions"
priority: P3
source: "meta AGENTS.md lines 533-592"
```

**Content** (generalized):

```markdown
## Knowledge Retrieval

Agents SHOULD prefer Dewey MCP tools over grep/glob/read
for cross-repo context, design decisions, and
architectural patterns. Dewey provides semantic search
across all indexed Markdown files, specs, and web
documentation — returning ranked results with provenance
metadata that grep cannot match.

### Tool Selection Matrix

| Query Intent | Dewey Tool | When to Use |
|-------------|-----------|-------------|
| Conceptual understanding | `dewey_semantic_search` | "How does X work?" |
| Keyword lookup | `dewey_search` | Known terms, FR numbers |
| Read specific page | `dewey_get_page` | Known document path |
| Relationship discovery | `dewey_find_connections` | "How are X and Y related?" |
| Similar documents | `dewey_similar` | "Find specs like this one" |
| Tag-based discovery | `dewey_find_by_tag` | "All pages tagged #decision" |
| Property queries | `dewey_query_properties` | "All specs with status: draft" |
| Filtered semantic | `dewey_semantic_search_filtered` | Semantic search within source type |
| Graph navigation | `dewey_traverse` | Dependency chain walking |

### When to Fall Back to grep/glob/read

Use direct file operations instead of Dewey when:
- **Dewey is unavailable** — MCP tools return errors or
  are not configured
- **Exact string matching is needed** — searching for a
  specific error message, variable name, or code pattern
- **Specific file path is known** — reading a file you
  already know the path to (use Read directly)
- **Binary/non-Markdown content** — Dewey indexes
  Markdown; use grep for Go source, JSON, YAML, etc.

### Graceful Degradation (3-Tier Pattern)

**Tier 3 (Full Dewey)** — semantic + structured search:
- `dewey_semantic_search` — natural language queries
- `dewey_search` — keyword queries
- `dewey_get_page`, `dewey_find_connections`,
  `dewey_traverse` — structured navigation
- `dewey_find_by_tag`, `dewey_query_properties` —
  metadata queries

**Tier 2 (Graph-only, no embedding model)** — structured
search only:
- `dewey_search` — keyword queries (no embeddings needed)
- `dewey_get_page`, `dewey_traverse`,
  `dewey_find_connections` — graph navigation
- `dewey_find_by_tag`, `dewey_query_properties` —
  metadata queries
- Semantic search unavailable — use exact keyword matches

**Tier 1 (No Dewey)** — direct file access:
- Use Read tool for direct file access
- Use Grep for keyword search across the codebase
- Use Glob for file pattern matching
```

## Injection Order

When multiple blocks are missing, they should be injected
in this order to maintain a logical document flow:

1. Core Mission (top of document, after Project Overview)
2. Gatekeeping Value Protection (inside Behavioral
   Constraints)
3. Workflow Phase Boundaries (inside Behavioral
   Constraints)
4. CI Parity Gate (inside Technical Guardrails or
   Behavioral Constraints)
5. Review Council PR Prerequisite (after constraints)
6. Spec-First Development (after constraints)
7. Website Documentation Sync Gate (near documentation
   gates)
8. Knowledge Retrieval (after coding conventions)

## No Persistent Data Model

This feature introduces no persistent data structures:
- No new Go structs
- No new JSON schemas
- No database tables
- No configuration files
- No state files

The guidance text is defined inline in the `/uf-init`
command file and injected directly into `AGENTS.md` at
runtime.
