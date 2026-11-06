---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:27Z
identity: documentation-gate-20261106T211227-jay-flowers
tier: draft
---

Cross-repo documentation gates can read as unsatisfied from issue titles and bodies while the scope extension lives in a comment: unbound-force/website#383's body tracked only PR 958 scope, while a comment extended it to PR 959's determinism scope. When verifying a documentation gate, read the tracking issue's comments (gh issue view --comments), not just the body; treat a comment-extension as satisfying the tracking requirement, and flag residual body/title staleness as a LOW advisory.
