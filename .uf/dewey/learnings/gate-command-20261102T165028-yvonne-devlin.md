---
tag: gate-command
author: yvonne-devlin
category: gotcha
created_at: 2026-11-02T16:50:28Z
identity: gate-command-20261102T165028-yvonne-devlin
tier: draft
---

When hardening a check or gate package with branch-aware artifact resolution, audit every check that globs artifact directories, not just the existence checks. In the uf-gate change (unbound-force PR #582), checkCoverageStrategy retained glob-all candidate collection (openspec/changes/*/tasks.md, specs/*/plan.md) after its sibling implement-phase checks adopted strict active-change resolution, so any sibling change's coverage language satisfied the Constitution IV coverage-strategy check for the active change — the exact false pass the branch-aware resolver was introduced to prevent. Detection pattern: grep the package for filepath.Glob over artifact base directories whenever introducing a stricter resolution path; every remaining glob site is a potential bypass of the new resolver.
