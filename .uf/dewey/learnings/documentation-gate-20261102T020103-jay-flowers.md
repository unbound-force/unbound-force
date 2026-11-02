---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:01:03Z
identity: documentation-gate-20261102T020103-jay-flowers
tier: draft
---

Docs-gate issues that track follow-up fixes can be silently partial-addressed by unrelated later changes — only a fresh read of the tracked artifact reveals it. Issue #747 (2026-10-17) tracked converting AGENTS.md's two stale "`/uf.unleash` Step 8" review-council references to the name-based form; the later unleash-tier-routing change fixed those same references via task 10.4 as a number correction ("Step 6, Code Review"), leaving the prescribed name-based form unapplied while the issue body still describes the pre-fix "Step 8" state. Two compounding staleness modes: absolute claims in the body age ("no references to /uf.run-pipeline-plan exist in docs/" no longer holds — docs/configuration.md:300 references it), and intervening changes renumber the anchored values the issue cites. Recipe: when triaging or executing a docs-gate follow-up, re-read the tracked artifact's current text and re-run the body's verification claims before acting; when a change fixes a tracked defect by different means, cross-reference the tracking issue so triagers can distinguish "done" from "done differently". Operational note: in this environment the Grep tool returned no matches for patterns verbatim-present on AGENTS.md:213 and :252 — verify AGENTS.md text with a direct read, not grep.
