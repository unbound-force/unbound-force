---
tag: ci-pinning
author: jay-flowers
category: gotcha
created_at: 2026-11-09T13:54:57Z
identity: ci-pinning-20261109T135457-jay-flowers
tier: draft
---

A SHA-pinned reusable workflow or action at an untagged commit cannot be repaired by adding a version comment: pinact --verify code 001 requires the comment's named tag to resolve to exactly the pinned SHA, and code 005 requires the comment to exist — together they force the pin to move to a tagged commit. When the pin has drifted to an untagged merge commit between releases, the compliant repair is a SHA-to-SHA repin to the nearest tagged commit whose referenced workflow file is byte-identical (verify line count and content SHA-256 via the GitHub contents API before editing), with the previous SHA recorded in the fixing PR as the documented rollback path. A comment-only refinement (per the earlier ci-workflows learning) works only when a tag already points at the pinned SHA. Verified in unbound-force PR #972: gaze pin 8bdfb59 (untagged, between v2.2.0 and v2.3.0) moved to the v2.3.0 tag commit 8031d419 with identical blob 86132290.
