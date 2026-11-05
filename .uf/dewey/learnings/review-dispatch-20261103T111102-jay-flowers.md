---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-03T11:11:02Z
identity: review-dispatch-20261103T111102-jay-flowers
tier: draft
---

When adding an input field to an OpenCode plugin tool that validates through a separate internal zod schema, the field must also be declared in the tool's host-facing `args` raw shape — the registration-surface seam means the host derives the model-facing parameter schema and validates tool input through z.object(args), so an undeclared field is invisible to calling models and can be stripped by host-side validation, silently disabling the feature. In the review-dispatch plugin, every schema-backed tool registers with `args: XxxInputSchema.shape`; consolidate_dispatch is the lone hand-duplicated inline args block and drifted the moment a field was added to its schema. Fix: register `args: ConsolidateDispatchInputSchema.shape` to structurally prevent recurrence.
