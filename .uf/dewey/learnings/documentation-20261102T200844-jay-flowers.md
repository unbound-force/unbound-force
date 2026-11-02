---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-02T20:08:44Z
identity: documentation-20261102T200844-jay-flowers
tier: draft
---

A scaffold-seeded config template whose commented example block is the primary field documentation is a user-facing docs surface that must enumerate every schema field. In preflight-project-checks, .uf/preflight.yaml's example block documented 6 of 7 fields (missing timeout_seconds, default 600 / max 3600) while docs/configuration.md and the zod schema carried all 7 — the copy-paste starting point was the incomplete surface. Review check: diff the template's commented fields against the schema's field list. Fixes must edit the canonical file and its scaffold mirror together (byte-identity is spec-pinned and drift-test-enforced).
