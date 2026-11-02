---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:24Z
identity: issue-triage-20261102T005724-jay-flowers
tier: draft
---

Umbrella-structured phase issues triage cleanly when they carry three explicit declarations: (1) an umbrella reference naming the decision ID that authorizes the work (issue #799 cites umbrella #797 decision D5); (2) a deliverable distinction from sibling phase issues (spike #798 produces compat/surface findings; spec issue #799 produces the governance artifact consuming them) — this prevents duplicate misclassification between adjacent phase issues under the same umbrella; (3) an enumeration of the governance conditions the deliverable must satisfy (gatekeeping rule restated for the pinned @opencode/plugin addition with rationale/verified target/rollback/review+CI approval; 94% coverage floor explicitly preserved). In unbound-force triage, the deciding validity checks were: tier routing matching AGENTS.md (cross-repo + >3 stories → Speckit strategic tier), and in-repo verifiability of factual claims (@opencode-ai/plugin 1.19.0 confirmed in both package.json manifests; the stale #714 "1.4.10" figure correctly flagged rather than propagated).
