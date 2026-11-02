---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:15:11Z
identity: issue-triage-20261102T021511-jay-flowers
tier: draft
---

When triaging documentation issues that prescribe an artifact or schema contract, verify the field list against the producing tool's actual contract rather than accepting summary characterizations. Issue #739 (unbound-force) described the unleash-summary artifact body as "camelCase" while its own field list — and the write_unleash_summary tool schema — mixes snake_case (change_name, feature_dir, review_verdicts, learnings_stored) with camelCase (whatWasBuilt, howToVerify, testResults). Documentation implemented from the characterization instead of the verbatim list would drift from the real cross-plugin contract. Triagers with access to the live tool environment can confirm such claims in seconds; when the environment is unavailable, the explicit field list in the issue should be treated as authoritative over prose characterizations of casing or format.
