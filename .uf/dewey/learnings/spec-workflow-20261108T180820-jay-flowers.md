---
tag: spec-workflow
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:20Z
identity: spec-workflow-20261108T180820-jay-flowers
tier: draft
---

A domain-codification change (constitution principle, convention pack) and its enforcement apparatus (reviewer persona, evidence tooling) are separate changes: when follow-on work grows from a specked change's review feedback — a new persona to own the new rules, new CLIs to feed it evidence — the prior change's completed tasks do not cover it. Before treating such follow-on work as specced, content-grep openspec/ and specs/ for the new artifacts' names and read the nearest prior change's proposal and tasks to confirm scope; if nothing matches, the follow-on needs its own spec tier and branch pattern. Tell-tale symptom: a CHANGELOG entry with no Spec: path in a file where every neighboring entry carries one. PR 959 (unbound-force) added the divisor-determinism persona and DE evidence tooling on a feat/* branch with no spec, while both its predecessors (instruction-lint, determinism-codification) were specced.
