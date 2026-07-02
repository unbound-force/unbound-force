---
tag: ci-causality-soft-gate
author: yvonne-devlin
category: pattern
created_at: 2026-07-01T10:06:25Z
identity: ci-causality-soft-gate-20260701T100625-yvonne-devlin
tier: draft
---

For AI agent instruction files (Markdown skills and commands), the two-tier baseline strategy (CI API first, local worktree fallback) provides good causality classification without doubling execution time. The CI API tier (gh api repos/{owner}/{repo}/commits/main/check-runs) is fast because it reuses data CI already computed. The worktree tier (git worktree add /tmp/preflight-baseline-SHORT_SHA main --detach) is the fallback for environments without gh or CI. A critical optimization is that only failing tools need baseline comparison — passing tools are not branch-caused by definition, so they skip baseline entirely. The conservative fallback (treat unknown as branch-caused) prevents false negatives at the cost of occasional false positives, matching /review-pr's established behavior.
