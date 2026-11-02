---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:51:01Z
identity: issue-triage-20261101T225101-jay-flowers
tier: draft
---

Presentation-layer changes to interactive command UX are cheap to govern when the tracking issue carries an explicit unchanged-gates clause naming the preserved gatekeeping surfaces. Verified triaging unbound-force/unbound-force#853 (question-tool-presentation-contract, 2026-11-01): the body states "Option sets, MANDATORY GATE markers, session-resume guards unchanged" alongside the presentation change (decision content printed as session text above the question tool, per-option reasoning, advisory recommendation, recommended option first, labeled (Recommended)). That clause lets the Guard close the gatekeeping-integrity check from the issue text alone — no diff inspection needed to confirm no workflow gate was weakened by the UX change. Combined with internally consistent counts (8+6+8=22 gates across /uf.triage-issue, uf.address-feedback, uf.review-pr) and full traceability (openspec/changes/question-tool-presentation-contract/, website twin unbound-force/website#339, originating issue #850), the issue triaged as VALID/enhancement/objective. Practice to encourage: presentation-contract issues should state their non-interference with option sets and gate markers explicitly, so governance triage stays objective without code access.
