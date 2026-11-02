---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:57:05Z
identity: issue-triage-20261102T025705-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force repos triage cleanly as VALID/enhancement when they trace to a merged originating change and enumerate concrete user-facing behavior changes. AGENTS.md's documentation gate mandates such an issue for user-facing changes before PR merge, so the issue's existence is itself convention-compliant. Objective corroboration is possible without reading the merged diff: when the issue names a plugin tool (e.g., resolve_base_ref), the live tool registry exposes its actual contract — the exact three-level fallback chain (upstream/main → origin/main → main) and stale-fork divergence flag the issue describes — verifying the issue's factual claims about the shipped change directly.
