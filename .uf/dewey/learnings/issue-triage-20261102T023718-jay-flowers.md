---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:18Z
identity: issue-triage-20261102T023718-jay-flowers
tier: draft
---

When triaging documentation-tracking issues ("docs: document X"), triangulate three independent evidence sources before ruling VALID: (1) the referenced change/spec artifacts exist and contain the claimed requirements — glob the change directory and read the requirement headings; (2) the underlying implementation is real — indexed PRs (via Dewey) and, for MCP-tool behavior claims, the tool descriptions and input schemas in the reviewing agent's own environment (e.g., consolidate_dispatch's dual-mode provenance description and finalize_review_dispatch's kind-specific verdict enums corroborate behavior claims verbatim without any code access); (3) the documentation gap is genuine — grep the docs tree for the key tool terms, noting that incidental mentions inside retrospective learnings reviews are not command-reference coverage. A documentation-gate issue that passes all three checks, carries checklist items traceable to spec FR numbers, and is labeled "documentation" satisfies the repo's documentation-gate rule for user-facing changes: rule it VALID, category enhancement, objectivity objective, no split — the constitutionally-mandated cross-repo website issue is correctly a separate linked artifact, not a split of the in-repo tracker.
