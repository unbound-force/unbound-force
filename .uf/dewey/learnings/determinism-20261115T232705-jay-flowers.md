---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T23:27:05Z
identity: determinism-20261115T232705-jay-flowers
tier: draft
as_of: "unbound-force@505dedf886fceeb2e552b7c1cf7633e2b6c36310"
---

The PR #993 adjudication pattern (mechanical prose + feasible extraction + unbound follow-up = CRITICAL DE-001) has a compliant judgment-prose complement, verified on PR #1129 (gaze.md and gaze-fix.md long-analysis notes, design Decision 4): guidance is DE-001-compliant judgment prose when all four conditions hold. (1) Its only procedure is an applicability decision — when to background a long analysis — with no prescribed mechanics: no loop, interval, completion check, or output format. (2) It self-declares the judgment: "This is an applicability judgment based on expected runtime and scope, not a required step." (3) The change record documents the owner lookup showing no registered tool performs the technique (closest: uf_finale_ci_poll, which polls PR CI, not local processes). (4) It carries a conditional migration clause: "If a launch-and-poll tool is later registered, migrate this guidance to invoke it." A conditional migration clause on judgment prose is forward-compatibility for future reviewers, not a deferred extraction obligation — the unbound-follow-up CRITICAL requires the shipped prose to BE mechanical.
