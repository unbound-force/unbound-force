---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:22:34Z
identity: issue-triage-20261102T052234-jay-flowers
tier: draft
---

Issues filed in a governance/meta repo that enumerate violations living in sibling repositories cannot be closed by any single-repo PR: GitHub issues are repo-scoped, and cross-repo fixes neither auto-close nor link to them. Triage of such issues should recommend splitting into per-repo issues (adding an audit task where the sibling-repo evidence is unenumerated) so each repo's own PRs and review councils can track and close the work; the meta-repo issue can then serve as an umbrella or be closed in favor of the per-repo issues. Corollary from issue #560 triage: line-number citations go stale when governing documents are amended after filing (constitution v1.4.0, amended 2026-10-14, shifted the Principle V supply-chain text from the cited lines 109-112 to 169-174), so triage should verify quoted text verbatim rather than trusting line anchors.
