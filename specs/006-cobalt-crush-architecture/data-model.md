# Data Model: Cobalt-Crush Architecture

**Spec**: 006-cobalt-crush-architecture
**Date**: 2026-03-20

## Entities

### Developer Persona Agent

A Markdown file defining the Cobalt-Crush AI developer
persona's identity, engineering philosophy, coding
checklists, feedback loop instructions, and speckit
integration behavior. Deployed to `.opencode/agents/`.

| Attribute | Type | Source | Notes |
|-----------|------|--------|-------|
| filename | string | File system | `cobalt-crush-dev.md` |
| description | string | YAML frontmatter | One-line persona description |
| mode | string | YAML frontmatter | "agent" (not subagent — runs as primary) |
| model | string | YAML frontmatter | AI model identifier |
| temperature | float | YAML frontmatter | Moderate (0.3-0.5 for creative coding) |
| role | Markdown section | H1 body | Engineering philosophy statement |
| source_documents | Markdown section | H2 body | Files to read before coding |
| engineering_philosophy | Markdown section | H2 body | Clean code, SOLID, TDD, CI/CD principles |
| code_checklist | Markdown section | H2 body | Convention pack adherence, test hooks, docs |
| gaze_feedback_loop | Markdown section | H2 body | How to read/address Gaze reports |
| divisor_preparation | Markdown section | H2 body | How to prepare for review |
| speckit_integration | Markdown section | H2 body | Task processing, phase checkpoints |
| decision_framework | Markdown section | H2 body | Ambiguity resolution, pattern selection |

**Identity**: Singleton — one `cobalt-crush-dev.md` per project.

**Lifecycle**: Created by `unbound init`. User-owned (never
overwritten without `--force`). Users may customize the
persona for their project.

### Convention Pack (shared — moved to `.opencode/unbound/packs/`)

Unchanged from Spec 005's data model. The only change is
the file location:
- **Old**: `.opencode/divisor/packs/{lang}.md`
- **New**: `.opencode/unbound/packs/{lang}.md`

Both Cobalt-Crush and The Divisor read from this shared
location. See Spec 005's data-model.md for the full
entity definition.

### Feedback Artifact

A structured output file produced by Gaze or The Divisor,
stored in `.unbound-force/artifacts/`. Cobalt-Crush reads
these to identify issues and apply learned patterns.

| Attribute | Type | Notes |
|-----------|------|-------|
| source_hero | string | "gaze" or "the-divisor" |
| artifact_type | string | "quality-report" or "review-verdict" |
| timestamp | ISO 8601 | When the artifact was produced |
| content | Markdown/JSON | Structured findings |
| path | string | `.unbound-force/artifacts/{type}/{timestamp}-{hero}.md` |

**Note**: The formal artifact schema is deferred to Spec
009 (Shared Data Model). Cobalt-Crush's instructions
reference the expected directory and file pattern, but
the agent should also read from common fallback locations
(stdout, `coverage.out`) when formal artifacts don't exist.

## Relationships

```text
Developer Persona Agent (cobalt-crush-dev.md)
       │
       │ reads
       ▼
Convention Pack (shared at .opencode/unbound/packs/)
       │
       │ also read by
       ▼
Divisor Persona Agents (divisor-*.md)

Developer Persona Agent
       │
       │ reads
       ▼
Feedback Artifact (.unbound-force/artifacts/)
       │
       │ produced by
       ▼
Gaze / The Divisor
```

## Validation Rules

1. `cobalt-crush-dev.md` MUST exist in `.opencode/agents/`
   after `unbound init` runs.
2. The agent MUST reference `.opencode/unbound/packs/` for
   convention pack loading (not `.opencode/divisor/packs/`).
3. The agent MUST function without Gaze or Divisor artifacts
   present (graceful degradation per Principle II).
4. Convention pack location MUST be identical in both
   `cobalt-crush-dev.md` and all `divisor-*.md` agents.
<!-- scaffolded by unbound vdev -->
