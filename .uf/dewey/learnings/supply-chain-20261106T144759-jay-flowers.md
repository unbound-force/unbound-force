---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-06T14:47:59Z
identity: supply-chain-20261106T144759-jay-flowers
tier: draft
---

Moving a SHA-pinned reusable workflow or action between two commits is provably zero-behavior-change when the referenced file's git blob SHA is identical at both pins: git blob hashes are content-addressed, so matching blobs guarantee byte-identical execution definitions regardless of how many commits separate the pins or what else those commits changed. In unbound-force PR #972 (fix-pinact-gaze-pin), the Gaze reusable_gaze_ai_report.yml pin moved across 16 upstream commits to reach the v2.3.0 tag commit, and blob identity (86132290e5334397f6d2d537e8a616ec1226e3bf at both SHAs) carried the zero-behavior-change claim past tooling bumps, a persona rename, and stderr secret redaction in the intervening commits. Prefer `git rev-parse <sha>:<path>` (or the GitHub contents API blob hash) for pin-move verification over line counts or text diffs — it is a single content-addressed comparison that cannot miss a difference.
