---
tag: determinism
author: jay-flowers
category: decision
created_at: 2026-11-24T14:55:48Z
identity: determinism-20261124T145548-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
---

Determinism spec-review scope gate for CI-only OpenSpec changes: an OpenSpec change whose Impact is confined to .github/workflows/*.yml edits and GitHub settings (ruleset via gh api) proposes no instruction files (.opencode/agents|commands|skills|uf/packs, AGENTS.md) and no tool code (.opencode/plugins, .opencode/lib), so Spec Review Mode produces no DE findings and APPROVES. Diff-based detection logic embedded in workflow YAML (e.g., a changes job's git-diff path classification) executes in the CI runner and satisfies Constitution Principle VI without tool extraction — the runner executes it, no agent re-reasons it from prose. Two adjacent candidates are correctly NOT findings: (1) duplication of such a block across workflows is DRY (Architect's domain, and the design explicitly weighed extraction); (2) a one-time post-merge ruleset-swap chain in tasks.md (gh api list → PATCH → read-back verify) is a standard operational migration task with judgment anchors (merge-relative timing, rollback-or-restore choice), not a DE-005 judgment-free instruction chain — DE-001/DE-005 govern recurring instruction text in commands/agents/skills. Evidence pattern: instruction-lint --changed returns delta.files: [] and tool-inventory --seams returns [] , which is the affirmative tool evidence for the scope gate.
