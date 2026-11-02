---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:42:35Z
identity: issue-triage-20261102T034235-jay-flowers
tier: draft
---

When triaging issues that refactor .opencode/agents/divisor-*.md files, verify the roster against two authoritative sources rather than the issue's own count. (1) Dispatch universe: .uf/review-matrix.yaml (mirrored at internal/scaffold/assets/uf/review-matrix.yaml) lists six review-capable agents — adversary, architect, curator, guard, sre, testing — for all four modes (code/spec/triage/feedback); the scaffold command-contract test asserts the same "six review roles", with envoy, herald, scribe as content-only agents that are discovered but never dispatched. "All 5 Divisor agents" claims undercount and typically omit divisor-architect, which carries the same boilerplate sections as the named five. (2) Scaffold mirroring: the agent files are canonical sources mirrored byte-identically into internal/scaffold/assets/opencode/agents/ and enforced by drift-detection plus content-assertion tests (curator must retain the Content Portability section and CP-001; guard must retain the Swarm-path spawn checklist parameters), and uf init --divisor deploys them to other repositories. Agent-file refactors must sync the embedded mirrors, and any new file the agents depend on (e.g. a shared divisor-common skill) must itself join the scaffold manifest — otherwise deployed agents reference a skill that does not exist in target repos.
