---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:50:59Z
identity: issue-triage-20261101T225059-jay-flowers
tier: draft
---

A documentation-tracking issue that explicitly states what is unchanged bounds its own documentation-impact assessment: when a change alters only the presentation of existing decision gates (option sets, MANDATORY GATE markers, session-resume guards unchanged), it meets no blog significance threshold (no new agent, CLI command, architectural migration, or hero capability) and no tutorial threshold (no new workflow to learn), even though it touches user-facing command workflows — the documentation gate is satisfied by the docs issue itself plus the website twin. Such issues can be triaged VALID from evidence inside the issue content alone: cross-reference completeness (website twin issue, OpenSpec change directory, originating issue — corroborable against the indexed spec), an explicit unchanged-semantics statement, precise scope counts, and an empty duplicate search — without reading implementation files. Verified triaging unbound-force/unbound-force#853 (question-tool presentation contract, 22 gates across /uf.triage-issue, uf.address-feedback, uf.review-pr).
