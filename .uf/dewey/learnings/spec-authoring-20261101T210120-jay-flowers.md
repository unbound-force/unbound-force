---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:01:20Z
identity: spec-authoring-20261101T210120-jay-flowers
tier: draft
---

Task-number cross-references in OpenSpec proposals drift when tasks.md phases are reorganized during authoring. In model-matrix-scaffold-deployment, the proposal cited "task 5.4" for the documentation-issue filing (header comment and the Impact "Documentation gate" bullet) while the final tasks.md numbered that task 6.1 inside a separate Bookkeeping phase — the change's notes.md recorded the correct number, proving the proposal simply was not re-checked. When a change's tasks.md gains, splits, or reorders phases after the proposal is drafted, grep the proposal for task-number references and re-verify each against the final tasks.md before the spec-review gate; stale references misdirect implementers and reviewers to nonexistent tasks. (Lesson proposal from divisor-curator, #889 review-council dispatch.)
