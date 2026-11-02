---
tag: command-contracts
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:53:35Z
identity: command-contracts-20261102T045335-jay-flowers
tier: draft
---

When a command-prose change removes a documented tool-call pattern (e.g. a no-provenance consolidate_dispatch retrieval) and introduces a replacement contract, content-gate tests must pin BOTH the new positive contract AND the absence of the removed pattern for EVERY requirement the change introduces — not only the requirements enumerated by the spec's test-mandate FR. Enumerating the test list inside a spec FR (e.g. FR-016 of review-pr-consolidate-dispatch) creates a floor that can silently omit sibling requirements (FR-014 single-source proposals/warnings shipped with zero pins while FR-013/FR-015 were pinned exhaustively), leaving removed prose reintroducible from git history with no failing check. A closing checklist — "for each removed sentence, is there an absence pin? for each new MUST sentence, is there a positive pin?" — closes the gap. The removed pattern's distinctive phrasing (e.g. "exactly once WITHOUT provenance") is the safest absence fragment because it cannot collide with surviving prose.
