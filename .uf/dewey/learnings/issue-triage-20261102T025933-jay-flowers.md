---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:59:33Z
identity: issue-triage-20261102T025933-jay-flowers
tier: draft
---

Triage verification recipe for unbound-force tool-extraction issues (proposals extracting deterministic command prose into tested plugin tools): (1) confirm the cited prose still exists — line ranges drift as sibling changes land (issue #675 cited uf.review-pr.md lines 829-993; the content now lives at ~692-818 after PR #745 landed), so grep for signature content like the branch-name pattern rather than trusting line numbers; (2) grep the repo for the proposed tool name to rule out an existing implementation; (3) validate against Constitution Principle VI / determinism pack DE-001 and DE-005 — the issue should name the deterministic procedure, preserve human confirmation gates in the command layer (DE-006 judgment boundary), and require registration in ALL mirrored copies (.opencode/plugins/ plus the internal/scaffold/assets/opencode/ mirror with drift tests); (4) sibling issues under the same parent initiative (e.g., #659 Phase 2 GitHub API Tools) propose distinct tools and are not duplicates despite shared keyword context.
