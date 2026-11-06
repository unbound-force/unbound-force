---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:48:54Z
identity: spec-governance-20261106T184854-jay-flowers
tier: draft
---

When a dependency security bump amends an unarchived OpenSpec delta that hardcodes versions or enumerates security controls, sweep the entire delta rather than editing only the pin sentence: scenario GIVEN examples that name the old version, and "X and Y overrides MUST remain" enumerations, must gain the new value and the new override. In stop-shipping-plugin-lock (PR #957), amending only "pin Zod 4.1.8 exactly" would leave the new zod override — the mechanism that actually forces nested copies to the fixed version — absent from the governed text, so a future change could drop it while remaining spec-literal. Pin-enforcement tests force the manifest side of a bump; only the delta text governs the control itself. A repo-wide grep for the old version string is the cheap completeness check for the sweep.
