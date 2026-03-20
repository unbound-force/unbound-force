# Data Model: The Divisor Architecture

**Spec**: 005-the-divisor-architecture
**Date**: 2026-03-19

## Entities

### Convention Pack

A structured Markdown document defining language-specific
review criteria. Deployed to `.opencode/divisor/packs/`
and loaded dynamically at review time by persona agents.

| Attribute | Type | Source | Notes |
|-----------|------|--------|-------|
| pack_id | string | YAML frontmatter | Language identifier (e.g., "go", "typescript", "default") |
| language | string | YAML frontmatter | Human-readable language name |
| framework | string (optional) | YAML frontmatter | Framework name if pack is framework-specific |
| version | semver string | YAML frontmatter | Pack version for tracking changes |
| coding_style | Markdown section | H2 body | Rules with CS-NNN identifiers |
| architectural_patterns | Markdown section | H2 body | Rules with AP-NNN identifiers |
| security_checks | Markdown section | H2 body | Rules with SC-NNN identifiers |
| testing_conventions | Markdown section | H2 body | Rules with TC-NNN identifiers |
| documentation_requirements | Markdown section | H2 body | Rules with DR-NNN identifiers |
| custom_rules | Markdown section | H2 body (in `-custom.md` file) | Rules with CR-NNN identifiers |

**File layout**:
- `.opencode/divisor/packs/{pack_id}.md` — canonical rules (tool-owned)
- `.opencode/divisor/packs/{pack_id}-custom.md` — project extensions (user-owned)

**Identity**: `pack_id` is unique. One canonical pack per
language. One custom extension file per canonical pack.

**Lifecycle**: Created by `unbound init`. Canonical pack
updated automatically on subsequent runs (tool-owned).
Custom pack never overwritten (user-owned).

### Persona Agent

A Markdown file defining a reviewer persona's identity,
focus areas, review checklists, output format, and
decision criteria. Deployed to `.opencode/agents/`.

| Attribute | Type | Source | Notes |
|-----------|------|--------|-------|
| filename | string | File system | `divisor-{function}.md` pattern |
| description | string | YAML frontmatter | One-line persona description |
| mode | "subagent" | YAML frontmatter | Always "subagent" |
| model | string | YAML frontmatter | AI model identifier |
| temperature | float | YAML frontmatter | Low (0.1) for deterministic review |
| role | Markdown section | H1 body | Persona identity and mission |
| source_documents | Markdown section | H2 body | Files to read before reviewing |
| code_review_checklist | Markdown section | H2 body | Universal + `[PACK]` tagged sections |
| spec_review_checklist | Markdown section | H2 body | Spec review mode sections |
| output_format | Markdown section | H2 body | Finding template |
| decision_criteria | Markdown section | H2 body | APPROVE / REQUEST CHANGES rules |

**Identity**: `filename` is unique. Discovery pattern:
`divisor-*.md` in `.opencode/agents/`.

**Canonical personas**: 5 ship by default — guard,
architect, adversary, sre, testing. Users MAY add custom
personas (e.g., `divisor-accessibility.md`).

**Lifecycle**: Created by `unbound init`. User-owned
(never overwritten without `--force`). Users may customize
persona behavior for their project.

### Review Council Command

A Markdown file defining the `/review-council` command's
orchestration logic. Deployed to `.opencode/command/`.

| Attribute | Type | Source | Notes |
|-----------|------|--------|-------|
| filename | string | File system | `review-council.md` |
| discovery_pattern | string | Body | `divisor-*.md` in `.opencode/agents/` |
| known_roles | table | Body | Reference table of canonical persona roles |
| code_review_mode | Markdown section | Body | Parallel delegation, CI replication, iteration |
| spec_review_mode | Markdown section | Body | Spec review with hybrid fix policy |
| iteration_max | int | Body | Default 3 |
| verdict_rules | Markdown section | Body | APPROVE / REQUEST CHANGES / ESCALATED |

**Identity**: Singleton — one `review-council.md` per
project.

**Lifecycle**: Created by `unbound init`. Tool-owned
(auto-updated when content changes). This ensures the
discovery pattern stays current.

### Persona Verdict

The structured output produced by a single persona during
a review session. Not persisted as a file — exists as
content within the review report.

| Attribute | Type | Notes |
|-----------|------|-------|
| persona | string | Name of the persona (e.g., "guard", "architect") |
| verdict | enum | APPROVE, REQUEST_CHANGES, or COMMENT |
| findings[] | array | List of Review Finding objects |
| summary | string | Brief narrative summary of the review |
| reviewed_at | timestamp | When the review was performed |
| iteration_number | int | Which iteration this verdict belongs to |

### Review Finding

A single issue identified during review. Part of a
Persona Verdict.

| Attribute | Type | Notes |
|-----------|------|-------|
| id | string | F-NNN identifier, unique within the review session |
| severity | enum | CRITICAL, HIGH (mapped from MAJOR), MEDIUM (mapped from MINOR), LOW (mapped from INFO) |
| category | string | Freeform category (e.g., "error-handling", "naming", "security") |
| file | string | File path where the issue was found |
| line | int (optional) | Line number if applicable |
| description | string | What the issue is and why it matters |
| recommendation | string | How to fix it |
| persona_source | string | Which persona found this issue |
| convention_rule | string (optional) | Convention pack rule ID if applicable (e.g., "CS-006") |

### Council Decision

The aggregate outcome of a review council session.
Produced by the `/review-council` command's verdict
aggregation logic.

| Attribute | Type | Notes |
|-----------|------|-------|
| decision | enum | APPROVED, CHANGES_REQUESTED, ESCALATED |
| discovered_personas[] | array of string | Personas found via `divisor-*.md` scan |
| absent_personas[] | array of string | Known canonical personas not found |
| persona_verdicts[] | array of Persona Verdict | One per discovered persona |
| iteration_count | int | Total iterations performed |
| unresolved_findings[] | array of Review Finding | Findings not resolved after max iterations |
| reviewed_at | timestamp | When the council session completed |
| convention_pack_used | string | Pack ID (e.g., "go") |
| pr_url | string (optional) | URL of the PR under review |

**State transitions**:
```text
        ┌──────────────┐
        │   RUNNING    │
        └──────┬───────┘
               │ all personas complete
               ▼
    ┌─────────────────────┐
    │ all APPROVE?        │
    │                     │
    │ YES ──► APPROVED    │
    │                     │
    │ NO ──► any REQUEST  │
    │        CHANGES?     │
    │                     │
    │  YES ──► ITERATION  │──► max reached? ──► ESCALATED
    │          (re-run     │
    │           failing    │
    │           personas)  │
    └─────────────────────┘
```

### Deployment Configuration

Settings used by `unbound init --divisor` to generate a
project-specific Divisor deployment. Not persisted — these
are runtime parameters.

| Attribute | Type | Notes |
|-----------|------|-------|
| target_dir | string | Root directory to scaffold into |
| divisor_only | bool | True when `--divisor` flag used |
| lang | string | Resolved language (explicit or auto-detected) |
| force | bool | Overwrite existing files |
| version | string | Unbound binary version for markers |

**Language detection markers**:

| Marker File | Resolved Language |
|-------------|------------------|
| `go.mod` | go |
| `tsconfig.json` | typescript |
| `package.json` | typescript |
| `pyproject.toml` | python |
| `Cargo.toml` | rust |
| (none found) | default |

## Relationships

```text
Convention Pack ──loaded-by──► Persona Agent
                               │
                               │ produces
                               ▼
                         Persona Verdict
                               │
                               │ contains
                               ▼
                         Review Finding
                               │
                               │ references (optional)
                               ▼
                         Convention Rule ID

Review Council Command
       │
       │ discovers
       ▼
  Persona Agent[]
       │
       │ aggregates
       ▼
  Council Decision
       │
       │ contains
       ▼
  Persona Verdict[]
```

## Validation Rules

1. Convention pack `pack_id` MUST match the filename
   stem (e.g., `go.md` has `pack_id: go`).
2. Convention pack rules MUST have unique identifiers
   within their section prefix (no duplicate CS-001).
3. Convention pack rules MUST use `[MUST]`, `[SHOULD]`,
   or `[MAY]` severity indicators.
4. Persona agent files MUST match the `divisor-*.md`
   naming pattern to be discovered.
5. Persona verdict MUST include at least one of:
   `APPROVE`, `REQUEST_CHANGES`, or `COMMENT`.
6. Council decision MUST be `APPROVED` only when zero
   discovered personas issued `REQUEST_CHANGES`.
7. `shouldDeployPack()` filter MUST deploy `{lang}.md`,
   `{lang}-custom.md`, `default.md`, and
   `default-custom.md` — no other packs.
<!-- scaffolded by unbound vdev -->
