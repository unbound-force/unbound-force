---
tag: agent-permissions
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:57:26Z
identity: agent-permissions-20261108T185726-jay-flowers
tier: draft
---

When a changeset both hardens a governance pattern in one component and adds a new sibling component, build the sibling from the hardened shape and write the verification gate against the invariant's class, not the exemplar. Observed on unbound-force PR #959: the same PR hardened divisor-linguist's bash allow-list (terminal metacharacter denies after every allow) and added divisor-determinism whose allow-list carried only the pre-hardening shape — undetected because both hardening tests read divisor-linguist.md by hardcoded path. A gate pinned to one named file cannot detect divergence in components the same PR adds. The repo's own enrollment lock (TestReviewerCapabilities_RepoLocalPersonasEnrolled, globbing .opencode/agents/) is the pattern to copy: glob the class (every persona whose bash block contains an allow) and assert the invariant on each member.
