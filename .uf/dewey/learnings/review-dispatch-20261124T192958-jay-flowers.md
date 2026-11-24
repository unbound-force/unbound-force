---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-24T19:29:58Z
identity: review-dispatch-20261124T192958-jay-flowers
tier: draft
as_of: "unbound-force@12bb5f8799eea38393959349f27354e1bb031a77"
---

Zero-scope Linguist reviews cannot record their zero-findings state through submit_review_findings: the tool schema enforces findings minItems >= 1 and rejects an empty array ("Too small: expected array to have >=1 items"). When no changed path is an instruction file (e.g., a diff of only .github/workflows/*.yml, CHANGELOG.md, and openspec/ spec artifacts), the Linguist correctly produces no findings and APPROVEs — the zero-findings state must then be stated explicitly in the final text response, never satisfied by fabricating a placeholder finding to fit the schema.
