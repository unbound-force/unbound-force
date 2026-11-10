---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-08T23:16:20Z
identity: review-dispatch-20261108T231620-jay-flowers
tier: draft
---

Verdict-severity invariants in the review-dispatch plugin (verified 2026-11-08 while reviewing advisory-approval-event, PR #990): an APPROVE WITH ADVISORIES dispatch verdict cannot carry CRITICAL or HIGH findings or advisories. Run votes are tool-derived from finding severities (runVerdict, index.ts:3994-4004, applied at 5041), any CRITICAL/HIGH finding forces REQUEST CHANGES (computeVerdict 3953-3959, calculatedWorkflowValue 2251-2252), and advisories are built only from the MEDIUM/LOW findings of an already-advisory verdict (buildAdvisories at index.ts:5080). Consequence when reviewing verdict-to-event mapping changes: severity-keyed guard rows evaluated against dispatch data are unreachable defense-in-depth, and the live protection for review-level HIGH content that never enters the artifact (e.g. PR-caused CI failures classified HIGH by the parent command per uf.review-pr.md:715) is the posting command's prose check over the composed review plus its fail-closed unknown-severity rule. When assessing such a guard, verify the check's operand — composed review prose versus the validated read_dispatch_artifact findings/advisories sections — and whether its operative noun ("finding") unambiguously covers advisories, since a future pipeline change that reintroduces HIGH advisories would make that prose check the only guard.
