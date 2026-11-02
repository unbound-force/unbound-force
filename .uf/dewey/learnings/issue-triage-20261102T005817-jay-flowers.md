---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:58:17Z
identity: issue-triage-20261102T005817-jay-flowers
tier: draft
---

Documentation-gate triage in the unbound-force meta repo: a change that adds fail-closed parameters to a workflow tool contract (uf_delegate_command's task_classification/skill/skill_absent_reason with SKILL_REFERENCE_REQUIRED/SKILL_NOT_FOUND envelopes) is user-facing for documentation-gate purposes even when the human-typed command surface (/uf.unleash [description]) is unchanged, because the workflow contract and reviewer checklists changed. A well-formed documentation-gate tracker issue is verifiable against four signals: mandatory per-file dispositions (CHANGELOG, AGENTS.md, README.md, docs/), a blog/tutorial reassessment against existing epic trackers, a documented cross-repo overlap check (this repo plus unbound-force/website), and closure criteria tied to PR merge or disposition resolution. Structurally similar doc-gate issues for different changes (e.g., #834 for spawn-subtask-classification-params vs #791 for delegation-skill-enforcement) are not duplicates — the originating change, not the issue shape, determines identity.
