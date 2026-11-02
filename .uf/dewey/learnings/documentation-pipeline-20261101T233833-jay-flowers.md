---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:38:33Z
identity: documentation-pipeline-20261101T233833-jay-flowers
tier: draft
---

Documentation-gate issues can have their premise invalidated when the behavior they were filed to describe changes again before the docs are written. Precedent: unbound-force/unbound-force#478 asked to document the constitution-only `.specify/` outcome of the uf init sentinel bug; the fix-init-speckit-sentinel change then restored full Speckit delegation, changing that behavior, so #827 was filed to supersede it (same pages, updated behavior) and #478 was closed. Supersession hygiene for documentation issues: the replacement issue names the superseded issue and states why the scope shifted (same pages, changed behavior), and the superseded issue is closed so the same pages are never tracked in two open issues. Corollary for triage: fix changes restore intended behavior rather than adding capability, so companion blog/tutorial issues are correctly absent while the documentation gate still applies (see learning content-pipeline-20261016T172845-jay-flowers).
