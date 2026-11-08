---
tag: openspec
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:55:38Z
identity: openspec-20261107T195538-jay-flowers
tier: draft
---

In a MODIFIED restatement delta, recompute the scenario-count claim in the numbering preamble after every edit. An "all N scenarios carry forward" statement goes stale the moment a superseded scenario is split into several, or review-driven scenarios are appended: count the delivered #### Scenario: headers before writing (or updating) the claim, and phrase replacements precisely ("23 carry forward verbatim; the collision scenario is replaced by three new fallback scenarios") rather than as blanket carry-forwards. A future archiver verifying scenario continuity — the archive-time backstop the delta's own design relies on — reads that count as the contract. Observed on PR 948's retroactive unleash delta: the preamble claimed twenty-four scenarios carry forward while the delta delivered twenty-six headers (one collision scenario replaced by three #934-driven scenarios).
