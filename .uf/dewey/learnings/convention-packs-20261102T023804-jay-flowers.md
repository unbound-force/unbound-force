---
tag: convention-packs
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:38:04Z
identity: convention-packs-20261102T023804-jay-flowers
tier: draft
---

Pack-listing drift is a recurring structural gap in the unbound-force meta repo: changes that add files under .opencode/uf/packs/ (ci-convention-pack, model-matrix-scaffold-deployment) have landed without updating all three listing surfaces — the AGENTS.md Convention Packs list, the docs/usage.md Customization pack listing, and the docs/architecture.md Pack Types section. Only AGENTS.md carries a formal exhaustiveness requirement (determinism-codification change: the list MUST include every pack file present in .opencode/uf/packs/); the docs surfaces have no drift guard, unlike embedded scaffold assets which carry byte-parity tests. When triaging or fixing pack-listing issues, always recount from the live directory (enumerate .opencode/uf/packs) rather than trusting counts stated in issues or docs — issue #719 predicted 15 packs, the directory now holds 18, and both docs surfaces list only 15 of 18 (meta-prompt pair missing; determinism.md tracked separately under the determinism docs issue).
