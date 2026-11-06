---
tag: ci-pinning
author: jay-flowers
category: pattern
created_at: 2026-11-06T14:48:08Z
identity: ci-pinning-20261106T144808-jay-flowers
tier: draft
---

Corrects the comment-only refinement guidance in learning/ci-integration-20261019T134427 and gaze-ai-report design D1 ("a later tag can be added as a comment refinement without moving the pin"): under pinact --verify this only works when the tag's commit exactly equals the pinned SHA (code 001). A pin at an untagged merge commit can never gain a compliant version comment — no tag resolves to it — so the only compliant repair is a SHA-to-SHA repin to the tagged release commit. Behavior-neutral repin procedure: (1) resolve the tag commit via gh api and confirm it exactly equals the target SHA; (2) confirm the referenced workflow file is byte-identical between old and new SHAs (git blob hash via the contents API) so the repin is provably behavior-neutral; (3) record the previous SHA in the fixing PR as the documented rollback path. Consequence for future pins: pin at tagged release commits from the start — an untagged merge-commit pin guarantees a future forced repin even when the workflow content never changed. Verified in unbound-force PR #972 (ci_local.yml gaze pin 8bdfb59 → 8031d419 # v2.3.0; blob 86132290e5334397f6d2d537e8a616ec1226e3bf identical at both SHAs).
