---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:58:31Z
identity: issue-triage-20261102T025831-jay-flowers
tier: draft
---

Triage of documentation-gap issues is objective when the issue cites its sources: verify (1) the cited spec/change directory exists with completed tasks, (2) the CHANGELOG entry exists, (3) the configuration surface ships in code and its scaffold mirror, and (4) the target docs contain zero references to the token — issue-keyword duplicate searches establish an issue is not a duplicate, not that the documentation gap is real. Applied to unbound-force/unbound-force #676 (review-matrix tier_caps): all four checks pass. Gotcha discovered during verification: change artifacts can disagree with shipped code on security-relevant details — the tier-gated-agent-count tasks.md item 2.3 specifies severity MEDIUM for the tier_caps.heavy security-sensitive advisory, while spec FR-006 and the shipped code (.opencode/plugins/review-dispatch/index.ts line 2026, severity: "HIGH") mandate HIGH so the advisory surfaces in the verdict reason per Constitution V. When documenting shipped behavior, the spec delta and the code are authoritative over tasks.md planning text; documentation should cite FR numbers and code constants, not task checkboxes.
