---
tag: meta-prompt
author: jay-flowers
category: gotcha
created_at: 2026-11-15T16:30:41Z
identity: meta-prompt-20261115T163041-jay-flowers
tier: draft
---

When a command step calls a tool with an argument whose value is produced by a LATER step's mapping table, name the derivation point at the call site. Precedent: /uf.review-pr Step 8a calls assess_prior_review_staleness with `verdict: the review event this run will post`, but the council-verdict-to-event mapping (including the self-authored PR remap to COMMENT) lives in Step 8c, after the call. An agent that passes the council verdict verbatim ('APPROVE WITH ADVISORIES', 'REQUEST CHANGES' with a space) fails the tool's zod enum; the structured error envelope then routes to the command's skip-the-gate degradation, silently disabling the very gate the call was added to enforce. Review symptom: a tool argument described only semantically, whose value is derived in a later subsection of the same command.
