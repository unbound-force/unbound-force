---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-20T17:06:01Z
identity: spec-authoring-20261120T170601-jay-flowers
tier: draft
as_of: "unbound-force@581834811fb8f5636c9a65c632b8d2d812b24fac"
---

When a spec prescribes a governance surface or a "recorded" decision, naming the pattern ("pause-and-exit", "the existing human-judgment gates") is not naming the mechanism. Before writing the scenario, verify the concrete surface exists in the load set: the exit variant, the command step that recognizes the finding type, and the field or artifact where the decision is recorded. Scenario AND-clauses like "the finding resolves only through the recorded governance decision" are unobservable — and "while unresolved" is unbounded — when no surface records that decision (cf. learning/meta-prompt-20261120T153544). In unbound-force specifically: per-finding override recording exists only in uf.review-council's Spec Review Mode gate; Code Review Mode's fix loop (steps 4-5) has no override record, so a CODE-mode governance finding needs its recording surface named in the spec or built by a task.
