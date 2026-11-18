---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T19:09:53Z
identity: determinism-20261116T190953-jay-flowers
tier: draft
as_of: "unbound-force@067fcc2c0cdadebd00928e76c1fe237973de59e8"
---

Instruction prose that names a tool's response fields can silently drift from the tool's actual contract when the tool omits a submitted judgment product from its structured response: uf.triage-issue Section 4.4 told agents to record a `title` from create_child_issue's response, but the registered tool returns only {repo, issue_number, issue_url, attempts} (verified via tool-inventory owner lookup, .opencode/plugins/uf-workflow/index.ts:3426) — the mismatch survived until human review caught it (issue #1056; fixed by PR #1169 re-sourcing the title from the step-1 confirmed child-issue list). Durable guards: (1) prefer tool contracts that echo submitted judgment products (e.g., the confirmed title passed to create_child_issue) in the response, so consuming prose records from one self-contained structured result and needs no cross-reference to earlier context; (2) when an instruction file names a tool's response fields, pin those field names in an output-shape (golden) seam test so prose-to-contract drift fails a test instead of reaching human review.
