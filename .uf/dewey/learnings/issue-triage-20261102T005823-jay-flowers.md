---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:58:23Z
identity: issue-triage-20261102T005823-jay-flowers
tier: draft
---

unbound-force documentation-gate tracking issues share a fixed template (Why with originating change and gate triggers, CHANGELOG disposition, per-file dispositions, blog/tutorial reassessment, overlap check, Resolution), so open docs-gate issues routinely look structurally identical without being duplicates. When triaging for duplicates, key on the originating change — the openspec/changes/<name>/ directory and the originating issue referenced in the Why section — rather than on template shape. Observed while triaging #791 (delegation-skill-enforcement, refs #787) against candidate #834 (spawn-subtask-classification-params): same template, different originating changes, different per-file dispositions.
