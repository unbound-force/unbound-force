---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-10-31T23:10:16Z
identity: content-pipeline-20261031T231016-jay-flowers
tier: draft
---

Spec changes that satisfy the Curator's documentation-gate and content-coverage review follow the model-matrix-delegation pattern (openspec/changes/model-matrix-delegation, iteration 3): (1) a dedicated "Docs and guidance" What-Changes item naming every documentation target (docs/configuration.md, CHANGELOG.md, AGENTS.md structural updates, command guidance); (2) explicit "Documentation gate" and "Content coverage" bullets in the proposal Impact section that name the AGENTS.md gate and the org constitution Cross-Repo Documentation rule, and cross-reference the issue-filing tasks with before-PR-merge timing; (3) a dedicated Documentation phase in tasks.md covering config docs, CHANGELOG, AGENTS.md structural edits, the in-repo docs issue, and the cross-repo website issue. Content-coverage dispositions (below blog threshold; tutorial satisfied by configuration reference docs) carry reasoning and land in a tracked artifact (the website issue), which satisfies the spec-review content checklist with no findings. Verification pattern: reviewer confirms the named target files exist in-repo and runs duplicate-issue searches before accepting the planned filings.
