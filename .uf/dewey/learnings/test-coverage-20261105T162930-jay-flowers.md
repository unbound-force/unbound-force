---
tag: test-coverage
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:29:30Z
identity: test-coverage-20261105T162930-jay-flowers
tier: draft
---

When a function and its test matrix are rewritten together, diff the function's full return-branch set against the new test cases before calling the matrix complete. Rewrites naturally gain tests for the NEW outcomes while carried-over defensive branches (e.g., a malformed-input Warn preserved from the pre-rewrite function) silently lose coverage. Evidence: PR 957's checkManifestInstallConsistency rewrite produced six dedicated cases for seven outcomes; the single carried-over malformed-manifest Warn branch was the untested one, caught by maintainer review rather than the suite. The install-side analogue of the same condition (unparseable node_modules package.json) was tested, showing the omission was rewrite-blindness, not deliberate exclusion.
