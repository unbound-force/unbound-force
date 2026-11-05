---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:33:27Z
identity: documentation-20261105T183327-jay-flowers
tier: draft
---

When documenting tool parameter ranges in a configuration reference, derive each row from the parameter's validation schema rather than copying bounds from adjacent parameters. In the delegation-timeout-budgets change (PR #950), docs/configuration.md stated a 60,000 ms minimum for uf_delegate_command's timeout by analogy with idle_timeout's enforced floor, but the schema is z.number().int().positive().max(...) with no minimum — the test suite successfully passes timeout: 5. Cross-check every documented range against the validator clause that enforces it. (From divisor-architect review of PR #950.)
