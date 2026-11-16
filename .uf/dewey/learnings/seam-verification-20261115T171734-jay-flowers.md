---
tag: seam-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-15T17:17:34Z
identity: seam-verification-20261115T171734-jay-flowers
tier: draft
---

When a DE-001 prose→tool extraction lands, its seam-verification suite must explicitly drive the new tool's error prologue — auth failure, empty-lookup, and each endpoint's read failure — because a mock harness's default fixtures silently inherit into every test. In PR #1077's assess-prior-review-staleness.test.ts, the module-level FIXTURES map never overrode "auth status" and never returned a URL-less "pr view", so the gh_auth_failed and empty-repository envelopes shipped with no regression lock even though the success path and two other error paths were tested. Rule of thumb: for every error branch in a new tool's fetch prologue, add a fixture-override test in the same increment (typescript.md TC-003, DE-007 seam verification).
