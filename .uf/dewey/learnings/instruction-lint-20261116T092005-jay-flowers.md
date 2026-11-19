---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-16T09:20:05Z
identity: instruction-lint-20261116T092005-jay-flowers
tier: draft
as_of: "unbound-force@e12549c22a8129c4415bd648ce9665b915f057e2"
---

IL-050 duplication findings under `--changed` scoping anchor at any near-verbatim paragraph in a changed file, including paragraphs the change never touched — pre-existing duplication gets re-anchored at the changed copy. Before reporting, check the anchor line against the diff hunks and confirm whether the counterpart copies are unchanged at base; if the duplication fully pre-exists, the changed-line confinement rule excludes it (drop with disclosure rather than re-anchor). PR #1145 example: the sibling-repository-references paragraph at uf.run-pipeline-plan.md:131 mirrors opsx-propose.md, speckit.plan.md, and speckit.specify.md from before the change; the scoped linter surfaced it only because the file was in the changed set, and the instruction-lint.ts segmentation change (skipping full-line HTML comments) did not affect it because no copy has an adjacent comment.
