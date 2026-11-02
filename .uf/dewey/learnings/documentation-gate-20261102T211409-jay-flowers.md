---
tag: documentation-gate
author: jay-flowers
category: decision
created_at: 2026-11-02T21:14:09Z
identity: documentation-gate-20261102T211409-jay-flowers
tier: draft
---

Restorative bug fixes that reinstate user-visible behavior do NOT qualify for the documentation-gate no-change exemption (that exemption covers internal refactoring, test-only, CI-only, and spec-artifact changes only). The fix-uf-finale-ci-poll-gh-schema change demonstrates the full compliant pattern for a restorative fix: (1) file BOTH an in-repo `docs` issue AND a `unbound-force/website` `docs` issue before the implementing PR merges; (2) thread both issue numbers into the CHANGELOG trailing reference alongside `Fixes: #N`; (3) record the docs disposition in the proposal's Impact section (task 5.3-style sub-items); (4) decline blog/tutorial opportunities with explicit rationale — a restorative change measured against the pre-break baseline is not a new capability. The crisp distinction is dual: restored behavior needs documenting (docs gate: YES) but only new capabilities need announcing (content pipeline: NO). Propagate the decline rationale into the filed issue bodies themselves ("no blog post or tutorial is requested") so downstream writers inherit the disposition.
