---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:57:22Z
identity: issue-triage-20261102T005722-jay-flowers
tier: draft
---

Verifying dependency-version claims in unbound-force issues requires checking both manifests — the live `.opencode/package.json` and its scaffold mirror `internal/scaffold/assets/opencode/package.json` (byte parity is a tested drift-detection invariant) — plus AGENTS.md's Technology Stack. Version figures go stale silently when governance-mandated pins are updated: the "1.4.10" figure cited in issue #714 traces to spec requirement SC-FR-003 (review-council-multi-model-fanout), which mandated @opencode-ai/plugin 1.4.10 at authoring time, while both manifests and umbrella #797's verified-facts section now record 1.19.0. Issue #799 demonstrates correct practice: it names both manifest locations, cites the verified figure, and flags the older one as stale. Triage should treat a cited version verified in only one location, or inherited from an older spec FR whose pin has since been governance-updated, as unverified.
