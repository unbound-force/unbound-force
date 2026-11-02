---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:28Z
identity: review-dispatch-20261102T035428-jay-flowers
tier: draft
---

Go's `internal/` layout convention collides with path-based security-sensitivity heuristics. In unbound-force (Go), all business logic lives under `internal/` (artifacts, config, coveragegate, doctor, gateway, homebrew, sandbox, scaffold, schemas, setup, textutil), so review-dispatch's `profileDiff()` predicate `/(^|\/)(cmd|internal)(\/|$)/` marks nearly every business-logic PR — including scaffold embedded markdown and schema tests — as security_sensitive, forcing the heavy review tier (`diffTier`: security_sensitive || lines > 300 || files > 10). Verified at .opencode/plugins/review-dispatch/index.ts lines 1195–1231 while triaging issue #656 (observed on PR #653). Design rule: security-sensitive path heuristics must key on security-relevant subsystem names (e.g., internal/gateway/, internal/sandbox/) or content keywords (secret|token|key|auth|credential|crypto), never on Go package-layout markers like `internal` or `cmd`. Narrowing must stay precise in both directions — under-classification silently routes security-relevant diffs to lighter review, a weakened gate with no audit trail.
