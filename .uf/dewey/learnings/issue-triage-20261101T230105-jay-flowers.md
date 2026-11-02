---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:01:05Z
identity: issue-triage-20261101T230105-jay-flowers
tier: draft
---

A documentation-gate issue that concludes "no update required" is a VALID compliance record, not a request for documentation work: the AGENTS.md documentation gate requires filing a documentation issue against the current repo for user-facing changes before PR merge, so the recorded impact assessment is itself the deliverable. Triage such issues by cross-verification rather than the body alone — verified on unbound-force/unbound-force #838 (update-run-pipeline-plan): (1) replicate the duplicate search to confirm no other open issue covers the command; (2) view referenced earlier gate-issues live — #747's comment carried the distinct AGENTS.md name-based-referencing follow-up that justified keeping it open while its superseded assessment was quoted accurately; (3) query Dewey for the change's own learnings — same-PR learnings corroborated the correction block's provenance (uf-workflow-20261026T182050's recorded delegation timeouts at question gates) and revealed the companion website gate issue (unbound-force/website#332) that the issue body never references, confirming the full in-repo + website gate pattern (content-pipeline-20261016T172845-jay-flowers) was followed; (4) the change's spec artifacts (proposal, design, command mirror) corroborate every enumerated behavior change. A dated correction block citing the falsifying PR review and learning ID strengthens the record: it documents falsification-and-repair rather than silently rewriting the claim.
