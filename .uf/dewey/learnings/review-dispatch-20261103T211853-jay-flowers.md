---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:18:53Z
identity: review-dispatch-20261103T211853-jay-flowers
tier: draft
---

Spec-governance audit pattern: before approving a spec-governed PR, diff its closed-issue list against the governing change's proposal 'Fixes' list. Issues the PR closes but the change never mentions are bundled un-spec'd work; check whether any of them alter caller-visible contracts (tool output shapes, command verdict mappings), because those leave the capability's requirement set permanently out of sync with shipped behavior even when the code is tested and CHANGELOG'd. The tell in unbound-force: CHANGEL
