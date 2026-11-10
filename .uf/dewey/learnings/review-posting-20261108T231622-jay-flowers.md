---
tag: review-posting
author: jay-flowers
category: pattern
created_at: 2026-11-08T23:16:22Z
identity: review-posting-20261108T231622-jay-flowers
tier: draft
---

When a change re-keys a mapping that other pipeline steps condition on — e.g. /uf.review-pr's council-verdict-to-GitHub-event mapping (PR #990, advisory-approval-event) — audit every sibling condition keyed to the same discriminator, not just the step being edited. PR #990 correctly re-keyed the Step 8c mapping table, the self-authored rule, and the confirmation-gate option sets to the mapped event, but left Step 8a's stale-dismissal and CODEOWNER warnings on "(APPROVE verdicts only)", which is evaluated on the pre-mapping council verdict — so newly-approvable advisory verdicts post APPROVE events without the warnings that exist to precede approvals. Method: grep the command (and its byte-identical scaffold mirror) for every occurrence of the discriminator's values (APPROVE, COMMENT, verdict) and classify each site as mapping, condition, or display before finalizing the change.
