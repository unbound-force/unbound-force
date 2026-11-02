---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:57:47Z
identity: issue-triage-20261101T225747-jay-flowers
tier: draft
---

Documentation-issue triage: duplicate candidates surfaced by fallback title-keyword search need content-genre analysis before a duplicate verdict. In unbound-force/unbound-force, the keyword "uf.rebase-pr" surfaced issue #856 (reference docs for the port-uf-rebase-pr command port) as a duplicate candidate for issue #855 (a scenario-driven conflicted-PR recovery tutorial), while the primary semantic duplicate search returned no matches. Reference documentation for a change and a task-oriented tutorial are complementary deliverables: cross-link them and sequence the tutorial after the underlying change lands, rather than deduplicating. An empty primary semantic search alongside a keyword-only match is a strong signal of genre collision, not duplication.
