---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-03T11:11:18Z
identity: test-quality-20261103T111118-jay-flowers
tier: draft
---

Regression-guard tests that assert against the production constant under test (`expect(x).toBeLessThan(SUMMARY_LIMITS.maxBytes)`) verify internal consistency, not contract compliance. When a spec pins a value (e.g., RD-FR-049's 40,960-byte budget and 15/15/10/5 caps), the regression test must assert the spec literal — otherwise raising the constant in production keeps every test green and the budget guard silently disappears. Rule of thumb: spec-pinned numbers get literal assertions; production constants may be used only for behavior that the spec does not pin.
