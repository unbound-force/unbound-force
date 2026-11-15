---
tag: documentation-triage
author: jay-flowers
category: context
created_at: 2026-11-10T20:28:28Z
identity: documentation-triage-20261110T202828-jay-flowers
tier: draft
---

In unbound-force/unbound-force, the AGENTS.md documentation gate (docs issue required for user-facing changes before PR merge) is practiced through per-change docs-labeled GitHub issues — e.g., #1034 (effect 4.0.0-beta.103 security bump for hero-repo adopters), #964 (lockless plugin install posture), #1012 (merge-plugin-runtime-deps, label "documentation") — while tactical follow-up changes are tracked by their originating issue (#991), which the PR closes. Curator reviews should therefore (a) keyword-search open docs-labeled issues AND the originating issue before reporting a missing documentation issue, and (b) treat security-override pin changes to the scaffolded .opencode/package.json as user-facing: hero-repo adopters pick the pins up through uf init, so they warrant their own docs issue. When a review-child session is prohibited from creating issues, carry the ready-to-run gh issue create command in the finding recommendation instead of filing.
