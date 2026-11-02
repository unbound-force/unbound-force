---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:57Z
identity: issue-triage-20261102T001557-jay-flowers
tier: draft
---

When triaging documentation-gate issues that catch user-facing docs up to a shipped behavior change, verify the gap direction before sizing the work: grep the target docs for BOTH stale descriptions of the old behavior AND absence of the new behavior. In unbound-force/unbound-force#808 (per-invocation dispatch correlation lifecycle, follow-up to #794/PR #809), the docs were silent rather than stale — README.md, docs/usage.md, and docs/architecture.md contained no text implying the old one-correlation-per-conversation binding — so the issue's "update stale text" checkbox was a verified no-op while the "add lifecycle documentation" checkboxes carried all the work. A silent-docs gap needs new sections; a stale-docs gap needs rewrites; conflating the two misestimates the task.
