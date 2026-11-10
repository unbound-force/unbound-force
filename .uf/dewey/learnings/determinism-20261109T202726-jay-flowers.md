---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:26Z
identity: determinism-20261109T202726-jay-flowers
tier: draft
---

When a prose→tool extraction removes shell interpolation for the tool's own git/gh calls but downstream command sections still render the tool's returned values into shell templates, the refname charset gate must move INTO the tool as a fail-closed error code rather than being dropped: the tool's success envelope becomes the gate's enforcement point, so every value that later reaches a template is validated in code before the tool returns it. Argument arrays protect only the tool's own invocations, not downstream prose templates. Shipped as design decision D3 of the checkout-pr-head change: UNSAFE_PUSH_TARGET (branch/merge ref outside A-Za-z0-9._/-, remote not a plain name or userinfo-free https URL) fires inside checkout_pr_head because Section 4.4 of uf.address-feedback.md still renders HEAD_REMOTE/HEAD_MERGE_REF into quoted shell templates.
