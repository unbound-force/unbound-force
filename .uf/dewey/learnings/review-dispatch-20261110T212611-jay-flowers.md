---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:26:11Z
identity: review-dispatch-20261110T212611-jay-flowers
tier: draft
---

When adding a post-assembly byte budget to a bounded tool result, audit every field the budget newly caps for a recovery channel. A flagged truncation still loses data on paths where no artifact is written: in the review-dispatch plugin, finalize failures return artifact_path null and read_dispatch_artifact has no errors section, so error text beyond a per-string cap is unrecoverable. Size caps against the strings the codebase itself composes on the capped path — formatValidationError (.opencode/lib/reviewer-manifest.ts) joins every Zod issue into one string that passes 256 characters with about three issues.
