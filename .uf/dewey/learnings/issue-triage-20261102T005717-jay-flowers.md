---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:17Z
identity: issue-triage-20261102T005717-jay-flowers
tier: draft
---

Well-formed CI-verification issues in phased migration umbrellas share four verifiable triage signals: (1) machine-checkable done criteria stated as CI outcomes (e.g., "both legs green on main; a plugin regression under either runtime fails CI"), (2) the legacy runtime leg kept green so gates are additive rather than weakened, (3) explicit gatekeeping routing for governance-controlled CI changes (landing through spec approval per Behavioral Rules), and (4) accurate references to existing CI infrastructure that can be verified against the repo — issue #800's ci_local.yml:108 opencode-config secret usage and make plugin-test target both check out exactly. Issue #800 (OpenCode v2 CI leg, umbrella #797 Phase 1) exemplifies the pattern: it complements manual-validation issue #714 with continuous verification, converting a point-in-time dual-compat claim into enforced regression protection (Constitution Principle III). When triaging such issues, verify these four signals against the repository and confirm the umbrella's phase tracking lists the issue as distinct from its manual-validation sibling before ruling out duplication.
