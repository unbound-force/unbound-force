---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:52Z
identity: issue-triage-20261102T023552-jay-flowers
tier: draft
---

Triage pattern for docs-tracking issues that reference a merged PR (verified on unbound-force #706, checkpoint/get_progress/clear_progress docs): accept as VALID after two in-repo checks. (1) Capability exists: plugin tool registration in .opencode/plugins/uf-workflow/index.ts (PROGRESS_DIR = ".uf/progress" at line 153, createClearProgressTool at line 485, tool-map registration at line 5334) plus tool invocation in the named command files (uf.unleash.md, uf.finale.md, uf.review-council.md, uf.address-feedback.md). (2) Docs gap still open: a complete repo-wide grep for the capability's user-visible artifact (`.uf/progress`, 64 matches) plus targeted get_progress/clear_progress greps in docs/, README.md, and AGENTS.md returned zero user-facing hits — only CHANGELOG entries, .gitignore, spec artifacts, and plugin source/tests. Duplicate assessment: keyword-adjacent issues from the same multi-increment modernization program (#739 = Increment 2 of #716, #757, #741) are related, not duplicated — compare referenced PR/change scope, not keyword overlap. Archived openspec change documents still describing the replaced execution-checklist pattern are immutable point-in-time records, not docs-gap targets.
