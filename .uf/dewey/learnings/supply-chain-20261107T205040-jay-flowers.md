---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:50:40Z
identity: supply-chain-20261107T205040-jay-flowers
tier: draft
---

When triaging a reviewer claim that a superseded hash is "recorded nowhere in-tree", grep the exact named digest and pin down WHICH historical state it refers to before accepting or refuting: a lock-record file can retain a different era's hash while the state the claim names has zero carriers. Verified on PR #948 feedback thread PRRT_kwDORVoNbc606RWt (uf-divisor, LOW supply-chain): the claim that the pre-bump canonical lock hash (1ac9cdf1…) was recorded nowhere in-tree held — both carriers (testdata baseline_lock_sha256 and the enforcing test const) were re-pointed to the current post-bump hash in the same diff — while the retained security_override_rollback.lock_sha256 (fef6f7f1…) is the pre-HARDENING state, a different era; citing it as "the pre-bump hash" is an era conflation that produces a false refutation in both directions. Corollary: git history plus a CHANGELOG revert note are legitimate carriers of a superseded state's rollback information, so the residual gap is record continuity at the re-point site (annotate the superseded digest and the revert commit beside the re-pointed value), not executability of the rollback.
