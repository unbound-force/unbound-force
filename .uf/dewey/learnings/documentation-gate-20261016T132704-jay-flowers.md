---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-10-16T13:27:04Z
identity: documentation-gate-20261016T132704-jay-flowers
tier: draft
---

OpenSpec changes in unbound-force/unbound-force with user-facing workflow impact must file TWO documentation issues before PR merge: (1) an in-repo issue per AGENTS.md's documentation gate, and (2) a cross-repo issue in unbound-force/website per the org constitution's Cross-Repo Documentation rule, which claims supremacy over project-specific guidance. Precedent: unleash-direct-review-dispatch tasks 6.2/6.3 and dispatch-finalize-determinism tasks 4.4/4.6 both dual-file. Spec artifacts that name only the in-repo issue under-file the gate, and a proposal's Constitution Alignment section will not catch the omission because it assesses only the six numbered principles, not the Development Workflow rules.
