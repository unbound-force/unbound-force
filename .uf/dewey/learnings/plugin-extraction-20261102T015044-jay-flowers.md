---
tag: plugin-extraction
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:50:44Z
identity: plugin-extraction-20261102T015044-jay-flowers
tier: draft
---

Plugin module extraction in unbound-force (verified while triaging #741, uf-workflow index decomposition): extracting cohesive modules from .opencode/plugins/*/index.ts into .opencode/lib/ follows the shipped 12-module lib precedent (branch-classification.ts, review-dispatch-artifact.ts, uf-workflow-types.ts, ...) and carries four per-extraction obligations the implementing change must hold: (1) a byte-identical scaffold mirror under internal/scaffold/assets/opencode/ (lib/ and plugins/ subtrees) enforced by scaffold_test.go drift detection; (2) registration in plugin_activation.go where applicable; (3) the 94% statements coverage floor on the .opencode/ vitest suite (governance-controlled); (4) an OpenSpec change, since moving code between modules requires a spec workflow per AGENTS.md. Refactor follow-up issues that cite provenance (review-council run ID + dispatch correlation ID), quantify structure claims (line counts, named symbols), and name the target extraction pattern triage as objectively VALID — every #741 claim verified at triage time (5,402-line index.ts, 27 tool({ registration sites, all named parser/artifact symbols present). Adjacent-scope issues targeting different files (#769, review-dispatch test helpers) are related, not duplicates; coordinate .opencode/lib/ module boundaries when both land.
