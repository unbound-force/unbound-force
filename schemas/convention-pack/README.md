# Convention Pack Schema

The convention pack schema defines the structural validation rules
for convention packs — Markdown files with YAML frontmatter that
configure coding conventions shared between Cobalt-Crush (developer)
and The Divisor (reviewer).

## Producer

Convention packs are authored by the **unbound** scaffold tool and
customized by developers.

## Consumers

- **Cobalt-Crush** — applies convention rules during implementation
- **The Divisor** — enforces convention rules during review

## Required Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `pack_id` | string | Unique pack identifier (e.g., `go`, `typescript`) |
| `language` | string | Target programming language |
| `version` | string | Pack version (semver) |

## Optional Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `framework` | string | Target framework (e.g., `cobra`, `react`) |

## Required H2 Sections

Every convention pack MUST contain these sections:

1. `## Coding Style` — formatting, naming, import rules
2. `## Architectural Patterns` — design patterns, package structure
3. `## Security Checks` — credential handling, path safety
4. `## Testing Conventions` — test framework, naming, isolation
5. `## Documentation Requirements` — comments, commit format
6. `## Custom Rules` — project-specific overrides

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-03-21 | Initial release |
