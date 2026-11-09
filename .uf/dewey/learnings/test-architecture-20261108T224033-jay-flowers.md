---
tag: test-architecture
author: jay-flowers
category: pattern
created_at: 2026-11-08T22:40:33Z
identity: test-architecture-20261108T224033-jay-flowers
tier: draft
---

Hermetic merge-base CLI testing pattern (from PR #959 tool-inventory/instruction-lint suites): when a CLI's behavior depends on merge-base git state (deltas, seam reports), build scratch git repositories under mkdtemp with injected Reader/Git seams, commit fixtures at the base using explicit `-c user.email/-c user.name` config (no global-config dependence), mutate the working tree for head-side scenarios, and cover the untracked-file input class (`git ls-files --others` paths) explicitly — local pre-flight flows routinely run with uncommitted+untracked trees, so untracked coverage is load-bearing, not an edge nicety.
