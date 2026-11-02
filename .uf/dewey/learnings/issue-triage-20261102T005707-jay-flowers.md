---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:07Z
identity: issue-triage-20261102T005707-jay-flowers
tier: draft
---

Triage signal for gate-governed CI-extension issues in unbound-force repos: a CI matrix-leg proposal is well-formed when it (a) explicitly routes landing through spec approval per the Gatekeeping behavioral rule, (b) is additive — keeps the existing leg running the same suite so no gate is weakened or bypassed, and (c) states blocking failure semantics as its acceptance criteria (e.g., "a plugin regression under either runtime fails CI" makes the new leg a gate, not an advisory). Supply-chain note for triaging such issues: an exact npm version pin (package@x.y.z) is registry-immutable — npm forbids republishing a released version — so it adequately pins a CI-installed CLI such as @opencode/cli; the commit-SHA pinning rule (CI-001) governs GitHub Actions `uses:` refs where mutable tags can be moved by publishers, not package-manager installs.
