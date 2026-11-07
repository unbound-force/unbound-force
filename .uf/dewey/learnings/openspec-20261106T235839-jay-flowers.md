---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-06T23:58:39Z
identity: openspec-20261106T235839-jay-flowers
tier: draft
---

OpenSpec tasks.md stable task identifiers (meta-prompt pack MP-016, N.N-style ids) apply to retroactive/spec-record changes too — there is no exemption for a change whose tasks are all already complete. `openspec validate --strict` validates delta structure but not task ids, so a retroactive change authored as a completion record silently drops the numbering that every executing change carries (delegation-skill-enforcement `1.4`, commands-dir-migration `1.1`-`10.5`, uf-gate `N.N`, adopt-org-infra `T N.N`), breaking the repo's task-citation practice (code comments cite "task 1.5", "tasks 5.1-5.4"). Reviewer check: scan checkbox lines in openspec/changes/*/tasks.md for a leading stable id (N.N, T N.N, or Step N) and compare against the nearest sibling change's format; flag bare-prose checkboxes as an MP-016 deviation regardless of completion state.
