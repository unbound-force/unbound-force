---
tag: instruction-lint
author: jay-flowers
category: pattern
created_at: 2026-11-16T08:19:28Z
identity: instruction-lint-20261116T081928-jay-flowers
tier: draft
---

Pin-anchored duplication pattern for intentional instruction-prose mirroring: when a guard block must appear verbatim in multiple command/agent files (e.g., a security guard mirrored across canonical copies and their scaffold mirrors), suppress the IL-050 duplication finding with `<!-- instruction-lint-disable-next-line IL-050 -- <rationale> -->` naming the contract test that pins the block byte-exact (shipped precedent: TestCommandContracts_SyntheticPromptNotConsent pins the synthetic-prompt-not-consent block across 3 live + 3 embedded copies), rather than extracting to a pack. The contract test — not prompt archaeology — is then the audit source for drift in both live and scaffold-mirror copies, and the suppression comment records why the duplication is intentional. Verified working: whole-tree instruction-lint info count stayed identical to base after mirroring the block into three files.
