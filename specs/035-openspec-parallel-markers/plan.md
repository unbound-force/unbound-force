---
title: OpenSpec Parallel Markers — Implementation Plan
status: draft
created: 2026-05-03
spec: "035"
---

# Implementation Plan: OpenSpec Parallel Markers

## Scope

Two files modified, one file updated:

1. **Task template** — Add `[P]` marker examples and
   a comment explaining when to use `[P]`
2. **Schema instructions** — Update the `tasks`
   artifact's `instruction` field to guide the LLM
   on when and how to add `[P]` markers
3. **AGENTS.md** — Document the `[P]` alignment between
   Speckit and OpenSpec task formats

## Approach

This is a template-and-documentation change. No Go code
is modified. No tests are affected (the template is a
Markdown file, the schema is YAML — neither is compiled
or tested by the Go test suite).

The `[P]` marker is already parsed by `/unleash` and
Replicator via string matching. No parser changes are
needed — the marker format is the same regardless of
whether the task ID is Speckit-style (`T001`) or
OpenSpec-style (`1.1`).

## Files

| File | Change |
|------|--------|
| `openspec/schemas/unbound-force/templates/tasks.md` | Add `[P]` examples + comment |
| `openspec/schemas/unbound-force/schema.yaml` | Expand `tasks` instruction |
| `AGENTS.md` | Document `[P]` alignment |

## Risks

None. This change is backward compatible — existing
`tasks.md` files without `[P]` markers continue to
work unchanged.
<!-- scaffolded by uf vdev -->
