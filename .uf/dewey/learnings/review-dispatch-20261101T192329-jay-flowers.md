---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-01T19:23:29Z
identity: review-dispatch-20261101T192329-jay-flowers
tier: draft
---

Council fix-loop advisories (follow-ups attached to an APPROVE WITH ADVISORIES verdict) have no enforcement mechanism — they live only in the PR body's prior-verdict summary and dispatch review context, and are silently dropped unless converted to tracking issues at advisory time. Observed on PR #898 (quick/accept-variantless-parent-model, 2026-11-01): of the advisory follow-ups listed in the PR body, the docs-issue verification was actioned (unbound-force/unbound-force#891 + unbound-force/website#360, both verified OPEN) while the governance-normalization issue for tier escalation was never filed — verified by searching all issue states for 'tier escalation', 'council-mandated', 'mid-flight', and 'quick tier opsx' with no match. Reviewer verification pattern: for each advisory listed in a PR's prior-council-verdict summary, search the repo's issues for a corresponding open tracker before treating the advisory as satisfied; when issuing an advisory that requires future work, either file the tracking issue in the same iteration or make the issue filing an explicit blocking task rather than leaving it in prose.
