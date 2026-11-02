---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:35:23Z
identity: issue-triage-20261102T033523-jay-flowers
tier: draft
---

Triaging tool-extraction issues in the unbound-force meta repo: when an issue proposing a skill (a prose protocol the LLM executes step-by-step) is re-scoped to an intent-level plugin tool, the determinism convention pack settles the direction without needing clarification — DE-005 mandates collapsing any judgment-free fixed tool-call sequence into a single intent-level tool, and DE-006 mandates that routing decisions (verdict-to-event mapping, severity sorting) execute inside the tool as constants or pure functions. Duplicate-handling consequence: a narrower sibling tool-extraction issue in the same pipeline (e.g., a standalone map_verdict tool next to a post_github_review tool) is absorbed as an internal function of the broader intent-level tool rather than shipping standalone, because a standalone lookup tool would leave the agent chaining calls — the exact pattern DE-005 removes. Triage should treat the broader issue as the surviving tracker and mark the narrower one superseded/absorbed, regardless of which was filed first.
