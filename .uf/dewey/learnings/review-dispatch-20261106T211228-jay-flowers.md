---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-06T21:12:28Z
identity: review-dispatch-20261106T211228-jay-flowers
tier: draft
---

In review-dispatch child sessions where issue creation is prohibited, the documentation-gate check degrades to read-only verification: detect the repository with gh repo view, list and view docs/blog/tutorial issues in both the current repository and the constitution's website repository (including issue comments), then report any gap as a finding carrying the ready-to-run gh issue create command for the developer to execute.
