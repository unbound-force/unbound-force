---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:42:47Z
identity: issue-triage-20261102T034247-jay-flowers
tier: draft
---

Triage scope check for issues that edit Divisor agent files (.opencode/agents/divisor-*.md) in unbound-force: the canonical files are embedded tool-owned scaffold assets — byte-mirrored copies live in internal/scaffold/assets/opencode/agents/ (scaffold.go special-cases the opencode/agents/divisor- prefix), scaffold_test.go drift-tests embedded vs canonical content (divisor-guard Swarm-path DE-008 checklist assertions, divisor-curator CP-001 assertions per issue #593, tool-owned lists), and uf init divisor-mode deploys them to hero repos (scaffold_stealth_test.go). An issue whose file scope names only the canonical .opencode/agents/ copies will fail make check on drift detection and leave hero repos stranded on old content; triage must require embedded-copy sync, scaffold manifest updates, and deployment of any newly referenced files (e.g., a shared skill) in scope. Also verify the agent inventory against the repo rather than the issue text: the dispatch contract test defines six review roles (divisor-adversary, divisor-architect, divisor-curator, divisor-guard, divisor-sre, divisor-testing) plus three content agents (envoy, herald, scribe) that share the Step 0/Source Documents boilerplate — issue counts citing "5 Divisor agents" are stale spec-019-era counts and undercount the affected files.
