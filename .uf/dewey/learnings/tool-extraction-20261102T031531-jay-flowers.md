---
tag: tool-extraction
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:15:31Z
identity: tool-extraction-20261102T031531-jay-flowers
tier: draft
---

Tool-extraction issues in unbound-force need a two-axis consumer verification before the tool is built. Axis 1 — footprint: grep the repo for the procedure's signature content, because the duplication typically extends beyond the files the issue names (issue #670 names uf.review-pr.md and uf.review-council.md for convention-pack loading, but the same deterministic pack-detection/enumeration prose also lives in uf.address-feedback.md §2.1 step 1, uf.agent-brief.md Section 9, and uf.cobalt-crush.md step 3; a consumer sweep limited to the named files leaves orphaned prose behind). Axis 2 — overlap: check whether an existing tool already performs the work for some consumers — build_review_prompts already reads convention packs from disk when assembling reviewer prompts, so the uf.review-pr.md §3.9 ladder (which ends by passing pack content to child agent prompts) may be redundant on the dispatch path, while uf.agent-brief.md's AGENTS.md pack-listing need is NOT covered by it. The correct outcome is either a tool scoped to the consumers with a live need, or closing the issue without building when the overlap is total. Building without both verifications risks an orphaned tool (zero-waste violation) or an incomplete extraction.
