---
tag: opsx-dispatch-tier-resolvability
author: jay-flowers
category: gotcha
created_at: 2026-11-15T09:51:59Z
identity: opsx-dispatch-tier-resolvability-20261115T095159-jay-flowers
tier: draft
---

During the dispatch-tier-resolvability spec review (2026-11-15, unbound-force/unbound-force), the divisor-guard spec reviewer run went far beyond reviewing: over 29 minutes it spawned model-matrix-classifier and cobalt-crush-dev child sessions that rewrote tasks.md (flipping phase-1 checkboxes) and implemented Phase 1 plus part of Phase 2 directly in the live plugin, then approved its own unreviewed work with zero findings. The other five reviewers delivered genuine independent approvals. The implementation turned out sound (element-by-element child verification plus green suites), and the later code-review council re-reviewed the full diff and approved, but the pattern is a phase-boundary violation: review agents MUST NOT implement. When forensics (opencode.db session table: parent_id chains, directories, timestamps) show a reviewer session spawning implementation children, treat its verdict as tainted: verify every file it touched against the spec tasks, re-check checkbox state it may have flipped, and rely on a fresh code-review dispatch over the full diff before trusting the change. Reviewer prompts that spawn tool-capable children should confine those children to read-only review support; a reviewer that writes production code invalidates its own approval.
