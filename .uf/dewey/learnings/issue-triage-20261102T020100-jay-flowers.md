---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:01:00Z
identity: issue-triage-20261102T020100-jay-flowers
tier: draft
---

Issue comments that cite edit targets by mutable anchors (step numbers, line numbers, section numbers) drift as the repository evolves — during triage, re-verify anchors against current repository state and locate edit targets by their descriptive position, not the cited literal. Verified instance (2026-11-01, unbound-force/unbound-force issue #747): the follow-up comment (2026-10-17) tracks two "AGENTS.md /uf.unleash Step 8 references" — the Review council behavioral rule and the PR Review Commands table; by triage time both sites read "Step 6, Code Review" after a renumbering, so a literal "Step 8" search returns nothing even though the tracked work (convert to name-based referencing) remains outstanding at AGENTS.md lines 213 and 252. The drift is itself the failure mode the tracked fix addresses: numbered cross-command references break on renumbering; name-based references do not.
