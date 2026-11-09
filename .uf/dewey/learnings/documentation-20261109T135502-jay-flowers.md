---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-09T13:55:02Z
identity: documentation-20261109T135502-jay-flowers
tier: draft
---

Curator triage pattern for CI-only pin repins (unbound-force/unbound-force): a one-line GitHub Actions `uses:` SHA repin whose referenced reusable-workflow file is byte-identical between old and new SHAs is a CI-only change. Even when accompanied by CHANGELOG.md and openspec/ change artifacts, no user-facing trigger path (cmd/, .opencode/agents|commands|skills/, internal/scaffold/, AGENTS.md, README.md, docs/heroes.md) is touched, so the documentation-issue gate is exempt (both AGENTS.md and the constitution's Cross-Repo Documentation rule list CI-only as exempt) and no blog/tutorial issue applies (no new agent, CLI command, or workflow pattern). CHANGELOG.md is neither a trigger path nor an exempt path in the classification heuristic — treat its entry as the fulfillment of the documentation gate and verify factual accuracy (tag-to-SHA resolution, byte-identity, which CI checks actually run on the PR) rather than flagging a gap. When a review-dispatch prompt carries prohibitions that forbid issue creation, report any would-be filing as a finding containing the full `gh issue create` command instead of filing it.
