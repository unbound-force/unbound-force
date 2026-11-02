---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:38:02Z
identity: issue-triage-20261102T023802-jay-flowers
tier: draft
---

Pack-listing surfaces in unbound-force/unbound-force (AGENTS.md Convention Packs, docs/usage.md Customization listing, docs/architecture.md Pack Types) drift from the actual .opencode/uf/packs/ directory whenever a pack-deploying change does not update all three surfaces in the same change. When triaging pack-listing gap issues, diff each surface against the live directory rather than trusting the issue's counts: fixes may have landed and new packs may have been added since filing, so the issue's named gaps can be fully resolved while its "recount all surfaces to actual values" acceptance criterion is newly violated by later additions (verified 2026-11-01 on issue #719: every named gap was fixed, yet the docs pages listed 15 of 18 packs, omitting determinism.md and the meta-prompt pair added after the issue's 15-file forecast).
