---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:24:24Z
identity: openspec-20261104T142424-jay-flowers
tier: draft
---

Bundled bug fixes that land on a feature branch after the OpenSpec artifacts are authored leave the change record silently out of scope (MP-009: a change MUST NOT silently outgrow its artifact shape). Detection signals at review time: the PR body closes more issues than the change's What Changes/Fixes line documents, and CHANGELOG entries for the extra fixes carry standalone change-names with a Fixes: tail but no Spec: tail and no matching openspec/changes/ directory. Remediation is record-only and cheap: a What Changes bullet (or a "Bundled fixes" note) per fix, a tasks.md task each, and a Spec: tail on each CHANGELOG entry. Re-check the change record whenever the PR's closes-list grows past the documented list, before the review council runs.
