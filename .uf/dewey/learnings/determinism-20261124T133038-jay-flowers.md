---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-24T13:30:38Z
identity: determinism-20261124T133038-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
---

Determinism spec-review scope boundary (precedent: openspec change ci-skip-learnings-only-prs, issue #1194): a spec that proposes only GitHub Actions workflow edits, repository-ruleset settings via gh api, and a CHANGELOG entry proposes no instruction files (.opencode/agents|commands|skills|uf/packs, AGENTS.md) and no tool code (.opencode/plugins|lib), so divisor-determinism Spec Review Mode exits with zero findings. DE-001's prose-vs-tool dichotomy attaches to procedures an agent re-reasons from instruction text; a CI changes-job's git diff and path-prefix check is executable configuration run by the CI runner — the correct classification, which the change's own Principle VI assessment states. DE-007 seam-verification tasks are likewise inapplicable to GitHub Actions jobs: they have no plugin registration surface, no documented output shape, and no write→read round-trip seams. Do not flag CI-orchestration specs for lacking seam tasks or owner-tool extraction; DRY questions about the duplicated detection blocks belong to the Architect.
