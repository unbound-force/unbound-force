---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:10Z
identity: issue-triage-20261102T041010-jay-flowers
tier: draft
---

Retro-agent issues proposing new AGENTS.md guidance can be corroborated beyond their cited PR evidence by checking whether the repository's own change history already enforces the convention by example. Issue #609 (add test-pattern reuse guidance to AGENTS.md Testing Conventions) cites PR #594's duplicate fence-extraction test as the triggering incident; the repo's history independently confirms the claim — openspec/changes/add-speckit-ref-assertion-test/tasks explicitly directs new content-assertion tests to be "modeled on TestGuardrailTemplates_CommandSpecificContent" with mustContain/mustNotContain tables, showing the table-driven guardrailBlock pattern is the established convention that AGENTS.md merely fails to document. Verdict signal: when proposed guidance restates a convention already enforced by example in prior tasks.md/spec artifacts but absent from AGENTS.md, the documentation gap is objective and the issue is a straightforwardly actionable enhancement. Implementation note for such issues: proposed bullets that embed line-number references (~6605) go stale — reference the struct or test by name only.
