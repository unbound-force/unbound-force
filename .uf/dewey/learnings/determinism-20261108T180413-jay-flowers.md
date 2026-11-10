---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:04:13Z
identity: determinism-20261108T180413-jay-flowers
tier: draft
---

"Judgment calls written as prose" is not a DE-001 exemption — it is the DE-006 boundary case. When a change's design.md defends mechanical evaluator prose by calling its thresholds and path lists judgment calls, the judgment was exercised once at design time; the runtime evaluation remains mechanical, and DE-006 requires those judgment products to execute inside the tool as constants. Feasibility is decided by suite precedent, not by the deferral note: fetch_pr_checks / fetch_pr_review_context in the review-dispatch plugin prove the extraction is feasible, so the infeasibility exemption (cf. learning/determinism-20261104T153538) cannot apply, and a risk note calling extraction a "natural follow-up" fails the tracked-deferral bar (cf. learning/issue-triage-20261101T225409) when no issue or tasks.md entry binds it. Adversary corollary: a review-coverage gate encoded as prose loses three properties a tool enforces — schema validation of untrusted compare/review API data, fail-closed behavior on every fetch failure (PR 993 Step 3.11 specified fail-closed only for the review fetch, leaving the compare fetch's failure path to agent improvisation), and a pinned sensitive-path matcher (prose "naming auth, tokens, secrets, or credentials" admits interpretation variance in a security gate). One extraction remediates the determinism violation and the resilience gap together.
