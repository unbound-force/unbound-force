---
tag: testing
author: jay-flowers
category: gotcha
created_at: 2026-10-06T00:45:50Z
identity: testing-20261006T004550-jay-flowers
tier: draft
---

When adding test cases inside an existing test file that uses nested describe() blocks (like review-dispatch.test.ts), be careful about describe block boundaries. Inserting a new describe block in the middle of an existing one can accidentally close the parent early, leaving subsequent test cases orphaned outside their intended describe scope. The divisor-architect agent specifically checks for this pattern. The fix is to verify the brace structure: ensure the new describe's closing }) doesn't close the parent, and that the parent's original closing }) still exists at the correct position. Use an editor or IDE's bracket matching to verify nesting before committing.
