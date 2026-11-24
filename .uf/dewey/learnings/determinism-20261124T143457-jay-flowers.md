---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-24T14:34:57Z
identity: determinism-20261124T143457-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
---

Spec-review determinism calibration for CI-orchestration changes (from ci-skip-learnings-only-prs iteration 2): a spec that assigns its deterministic procedures to GitHub Actions machinery — a `changes` detection job, `needs:`/`if:` gating expressions, an `if: always()` aggregate job — satisfies Principle VI by construction: the procedures run as versioned executable code in CI and never as agent prose, so DE-001/DE-003/DE-005 do not attach even though no plugin tool owns the work. Evidence pattern: instruction-lint delta.files empty, tool-inventory seams empty, owner lookups ("classify changed files", "ruleset required checks") return no owner. The review verifies the assignment — the proposal's Principle VI section naming the executable mechanism — rather than demanding a repo tool; a design decision rejecting extraction to a shared script/action/CLI (Security by Default dependency-necessity) is documented judgment that correctly stays in prose. One-time governance migrations written as exact gh api commands with judgment-interleaved identification, read-back verification, and mandated human surfacing of failures are not DE-005 judgment-free chains: DE-005 requires no judgment between steps, and human interaction is judgment territory.
