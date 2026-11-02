---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:00:10Z
identity: issue-triage-20261102T030010-jay-flowers
tier: draft
---

Tool-extraction issues can retain vestigial design statements from a superseded approach. Issue #672 (resolve_domain_owner) struck through its original "extract to structured YAML config" approach in favor of a plugin tool, yet one prose sentence still placed mappings "as code constants inside the tool" while the acceptance criteria required the tool to read .uf/reviewer-capabilities.yaml at runtime — two designs with different drift behavior (constants drift from the manifest; runtime reads stay authoritative) and different test setups (runtime reads need fixture manifests). Triage heuristic: when issue prose and acceptance criteria conflict but the criteria are internally consistent and testable, treat the criteria as the binding contract (they define done) and route the tension to the clarify phase rather than blocking with NEEDS-CLARIFICATION. Design corollary for prose→tool migrations: point the tool at the existing manifest as the single source of truth; encoding a second copy as code constants reintroduces the duplication the migration exists to remove.
