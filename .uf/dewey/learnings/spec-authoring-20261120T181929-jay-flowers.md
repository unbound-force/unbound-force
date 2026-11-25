---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-20T18:19:29Z
identity: spec-authoring-20261120T181929-jay-flowers
tier: draft
as_of: "unbound-force@581834811fb8f5636c9a65c632b8d2d812b24fac"
evidence:
  - "dewey:.uf/artifacts/dispatch/2c92d63f-6aed-45bd-aafd-01d3a6939506.json"
---

When an OpenSpec change declares a design decision's state fields normative ("the plugin tests pin their exact JSON serialization") while its spec delta mandates outcome enums in MUST requirements, the two enum value sets must be diffed literally at spec-review time — including "the same result enum as X" phrases, which silently go stale when one side gains a value the other lacks. Failure mode observed reviewing intent-satisfaction-check (2026-11-20): design D2 defined general-intent outcomes as using "the same result enum" as per-criterion outcomes (satisfied | unsatisfied | deviation) while RC-FR-012 mandated satisfied | unsatisfied | unevaluable; tasks pinned D2's shapes in plugin tests, making the requirement unimplementable as written. Reviewer checklist: for every enum that appears in both a design decision and a requirement of the same change, enumerate both value lists side by side and confirm set equality.
