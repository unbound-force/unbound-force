---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:07:19Z
identity: documentation-pipeline-20261107T160719-jay-flowers
tier: draft
---

A documentation-gate issue that lists surfaces as "included in PR #N" makes a falsifiable claim: cross-check every listed surface against the PR's changed-files list before merge, not just the issue's existence and labels. Verified on unbound-force PR #958: gate issue #970 listed docs/heroes.md ("10-persona roster, linguist repo-local note"), docs/usage.md ("persona counts and linter usage"), and an AGENTS.md "instruction-lint drift-exemption bullet" as included in PR #958, but the 30-file diff contained neither docs file — heroes.md still read "9-persona review council" (line 113), usage.md still said "divisor-* (9 agents)" (line 35), and no docs/ file mentioned lint-instructions — while AGENTS.md's diff carried only build-command, structure, and stack edits. The openspec proposal's own Impact section was accurate (CHANGELOG + AGENTS.md only); only the gate issue overclaimed. This extends the label/body verification protocol (learnings documentation-pipeline-20261106T180904, content-pipeline-20261102T170225) with a third check: changed-files cross-verification of every "included" surface claim, cheapest via the review dispatch's changed-files manifest or git diff --name-only compared against the issue's surface enumeration.
