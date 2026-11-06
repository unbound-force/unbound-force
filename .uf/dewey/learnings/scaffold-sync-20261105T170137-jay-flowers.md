---
tag: scaffold-sync
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:01:37Z
identity: scaffold-sync-20261105T170137-jay-flowers
tier: draft
---

Scaffold-mirror byte-identity is verifiable from a manifest-delivered diff without reading both files in full: each file section's 'index <oldhash>..<newhash>' line carries git blob hashes, and when the canonical .opencode/ file and its internal/scaffold/assets/opencode/ mirror show the same post-image hash, the two blobs are byte-identical by definition. Verified on PR 950: all four mirrored pairs (uf.run-pipeline-plan.md 3d2157c3, agent-executor.ts 2199b0be, delegation-budget.ts 77f862da, uf-workflow/inde
