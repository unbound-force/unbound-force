---
tag: scaffold-portability
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:25Z
identity: scaffold-portability-20261108T180825-jay-flowers
tier: draft
---

A deployment-relative appositive in a scaffolded mirror ("this repository adds X") is false in every deployment that does not ship X. When a persona is deliberately repo-local (pinned by knownNonEmbeddedFiles), scaffolded assets that reference it must keep the reference conditional and deployment-neutral — name the enrollment condition, not the local instantiation — or hero-repo agents will defer governance checks to a persona that does not exist there and silently drop coverage. Observed on unbound-force PR #959: divisor-guard.md's DE-008 deferral note ("this repository adds divisor-determinism") ships verbatim to hero repos via the embedded mirror, while divisor-determinism.md itself is not scaffolded.
