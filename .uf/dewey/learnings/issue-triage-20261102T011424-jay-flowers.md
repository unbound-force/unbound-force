---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:14:24Z
identity: issue-triage-20261102T011424-jay-flowers
tier: draft
---

In unbound-force repos, documentation-gate umbrella issues (docs label, "docs: <change>" title, filed by a MEMBER before PR merge per the AGENTS.md documentation gate) are valid by construction once the underlying change touches user-facing paths (.opencode/commands/, .opencode/skills/, cmd/, internal/scaffold/, AGENTS.md, README.md, docs/). When triaging such an issue, a later-numbered issue reporting a narrow gap that overlaps the umbrella's scope (unbound-force #826 "risk floors missing from docs/usage.md" vs the earlier #772 tier-routing docs umbrella) is a residual-gap follow-up after the umbrella's delivery, not a duplicate: the umbrella's delivery landed (AGENTS.md quick-tier row; website PR #315 closing website#314 and citing "tracked by #772") while the narrower specifics remained missing. Keep the earlier umbrella as the change's documentation record, let the follow-up track outstanding work, and verify delivery state from the umbrella's comments and cross-repo PR references before consolidating or closing either.
