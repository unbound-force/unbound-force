---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:09Z
identity: opencode-v2-migration-20261102T005709-jay-flowers
tier: draft
---

CI dual-runtime verification pattern for the OpenCode v2 migration (umbrella #797, surfaced triaging unbound-force#800 Phase 1): v1 and v2 cannot be installed side by side — both claim the `opencode` command (umbrella-verified fact, cf. learning opencode-v2-migration-20261102T003408-jay-flowers) — so continuous dual-compat verification must use separate matrix legs, each installing exactly one runtime in an isolated job; a single CI job cannot exercise both runtimes. Complementary constraints for the implementing spec: (1) the e2e /uf.review-council smoke depends on OPENCODE_CONFIG_CONTENT, which per the existing gaze-ai-report precedent (ci_local.yml push-to-main gating) is unavailable on fork PRs — "both legs green on main" is the achievable done-when framing, and fork-PR behavior needs explicit skip-with-notice handling (the gaze reusable-workflow pattern); (2) the exact @opencode/cli version pin matches the umbrella risk register's mitigation for v2 pin churn ("exact pin resolved at work start").
