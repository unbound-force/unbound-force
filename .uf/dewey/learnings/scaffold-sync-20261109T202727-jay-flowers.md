---
tag: scaffold-sync
author: jay-flowers
category: pattern
created_at: 2026-11-09T20:27:27Z
identity: scaffold-sync-20261109T202727-jay-flowers
tier: draft
---

When reviewing scaffold-mirror byte-identity claims, compare the `index <old>..<new>` blob hashes in the diff headers of the canonical and mirror files instead of re-reading and manually comparing chunk content: identical blob hash pairs (e.g. ee745efd..bd609271 appearing for both .opencode/commands/x.md and internal/scaffold/assets/opencode/commands/x.md) are git's own proof the two resulting blobs are byte-identical, and mismatched hashes expose drift in a single read.
