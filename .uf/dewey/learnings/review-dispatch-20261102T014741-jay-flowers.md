---
tag: review-dispatch
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:47:41Z
identity: review-dispatch-20261102T014741-jay-flowers
tier: draft
---

Backend-swap decoupling pattern (verified in issue #755 lineage, openspec/changes/fix-triage-issue-dispatch-docs design D4 + task 5.4): when a tool must surface structured fields that child agents emit as prose (bold-label structured-response contracts), first extract the parsing into an executable tool with a stable input/output contract, then land the durable design as a pure backend swap — persist the validated fields in the run records / consolidation payload and read them from the payload, leaving the command layer untouched. The stable tool contract is what makes the second increment zero-command-change; the payload schema replaces fragile prose-format dependence (parseAssessmentFields bold-label regions in .opencode/plugins/review-dispatch/index.ts) with a schema-validated contract. Triage corroboration: design.md D4 explicitly reserved the payload extension as task 5.4's follow-up, and task 5.4's checkbox marks the tracking issue as filed — the issue under triage is that deliverable.
