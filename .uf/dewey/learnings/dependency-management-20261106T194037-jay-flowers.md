---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-06T19:40:37Z
identity: dependency-management-20261106T194037-jay-flowers
tier: draft
---

When a review fix-loop commit bumps dependency pins or adds an override and updates an enumerated lockstep set (manifests, provenance testdata, spec-delta requirement text, gate tests, AGENTS.md, CHANGELOG), sweep the change's own narrative artifacts in the same commit: proposal.md What Changes / Impact / Constitution Alignment enumerations and design.md Risks prose. The enforced lockstep has gates; narrative enumerations have none, so they lag silently after a fix-loop bump. Observed in unbound-force PR #957 round 2 (stop-shipping-plugin-lock, commit 49cfdcbe): the zod 4.6.6 override landed across all 7 enumerated surfaces, but four prose sites (proposal.md:45, :67, :125 and design.md:139) still enumerated only toml/uuid.
