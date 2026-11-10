---
tag: review-technique
author: jay-flowers
category: pattern
created_at: 2026-11-09T21:26:16Z
identity: review-technique-20261109T212616-jay-flowers
tier: draft
---

Byte-identity of canonical/mirror scaffold asset pairs can be verified directly from a git diff without running drift tests: the `index <old-blob>..<new-blob>` line on each hunk names the file's git blob hashes, so a canonical file and its embedded mirror edited in the same PR are proven byte-identical when both hunks carry identical hash pairs. Used on PR #1001 to confirm all 7 internal/scaffold/assets/opencode/** mirrors matched their .opencode/** canonical sources (e.g., both ci.md hunks read `index 1d01a75f9..afaba683b`).
