---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:11Z
identity: issue-triage-20261102T005711-jay-flowers
tier: draft
---

Umbrella-migration triage: an issue claiming to be phase N of an umbrella can be validated against the umbrella's own phase-tracking checklist. When the umbrella's task list names the issue by number in its phase (e.g., unbound-force #797 Phase 1 listing #714 "plugin port" and #800 "CI v2 leg" as distinct line items), the relationship is deliberate decomposition, not duplication, and the issue's complementary-scope claim ("complements #714, whose validation plan is manual") is confirmed by the sibling issue's scope. The umbrella's verified-facts section also supplies constraints that explain the phase issue's shape: v1 and v2 OpenCode cannot be installed side by side (both claim the `opencode` command), so dual-runtime verification must run as a separate CI matrix leg — the exact design #800 proposes — and the v2 version pin is flagged for re-resolution when work starts, which is why the issue leaves the exact version to the workflow. Verified triaging unbound-force issue #800 (2026-11-01): Dewey's github source index surfaced the #797 phase-tracking and verified-facts blocks, the primary keyword search returned no duplicate candidates, and every artifact the issue names (ci_local.yml, the opencode-config secret at ~line 108, make plugin-test, /uf.review-council) exists in repo documentation.
