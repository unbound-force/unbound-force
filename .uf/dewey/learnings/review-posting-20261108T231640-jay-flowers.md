---
tag: review-posting
author: jay-flowers
category: gotcha
created_at: 2026-11-08T23:16:40Z
identity: review-posting-20261108T231640-jay-flowers
tier: draft
---

When an OpenSpec change justifies a non-goal with a claim about a sibling command's behavior, verify the claim against the sibling's actual file structure before accepting it. In unbound-force PR #990 (advisory-approval-event), the proposal claimed "/uf.review-council is unchanged: its advisory verdict applies to spec review, where this problem does not arise" — but /uf.review-council's GitHub posting (Step 7) exists only in Code Review Mode (uf.review-council.md:818-825), its consolidation yields APPROVE WITH ADVISORIES for MEDIUM/LOW findings in code mode (:386-387), and Spec Review Mode has no posting at all; the claim inverted the sibling's structure. The error propagated from review-council-github-posting design D6's "Spec mode only" table note, which already contradicted that change's own code-mode-only posting scope — an error chain that survives because each generation copies the prior design note instead of re-reading the command.
