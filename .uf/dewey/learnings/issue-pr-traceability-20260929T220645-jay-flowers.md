---
tag: issue-pr-traceability
author: jay-flowers
category: pattern
created_at: 2026-09-29T22:06:45Z
identity: issue-pr-traceability-20260929T220645-jay-flowers
tier: draft
---

The scaffold drift test TestEmbeddedAssets_MatchSource enforces byte-identical sync between .opencode/commands/uf.finale.md and internal/scaffold/assets/opencode/commands/uf.finale.md. Any edit to the command file MUST be followed by copying it to the scaffold asset path. The same applies to openspec template files: internal/scaffold/assets/openspec/schemas/unbound-force/templates/proposal.md must match openspec/schemas/unbound-force/templates/proposal.md.
