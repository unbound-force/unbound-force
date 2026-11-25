---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-25T18:26:08Z
identity: review-dispatch-20261125T182608-jay-flowers
tier: draft
as_of: "unbound-force@59d260231240f979d60ee4fc3d451fa5e5c39723"
---

Review-council manifest prompts say "every file is in scope," but each persona's own definition narrows scope first — for divisor-linguist Code Review Mode the scope is the instruction-files path table (.opencode/agents|commands|skills|uf/packs markdown, AGENTS.md, internal/scaffold/assets/opencode/**/*.md mirrors). A PR touching only .opencode/lib|plugins|test TypeScript, Go source, CHANGELOG.md, and openspec/changes/** has zero in-scope files: skip the instruction linter (the pre-flight advisory also reports empty scope when the local checkout sits at base with the PR head unchecked-out), spot-retrieve diff chunks only to verify manifest fidelity and the nature of instruction-adjacent files (openspec tasks.md/proposal.md are spec artifacts, not instruction files), submit an explicit empty findings array rather than omitting the submission, and still emit the **Files Reviewed** attestation so the dispatch run avoids the empty_coverage_attestation warning.
