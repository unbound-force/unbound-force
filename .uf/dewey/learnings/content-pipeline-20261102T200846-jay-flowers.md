---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-02T20:08:46Z
identity: content-pipeline-20261102T200846-jay-flowers
tier: draft
---

New project-owned scaffold extension points in unbound-force warrant tutorial tracking beyond the dual docs gate (in-repo docs-labeled issue plus cross-repo unbound-force/website issue before PR merge; precedent #916 + website#369, #899 + website#362). Adopting an extension point requires learning a strict schema plus semantics — scope filtering, gate effects, soft-gate baseline contract — that a configuration-reference row does not teach end-to-end. Precedent: reviewer-capabilities-init-survival filed a combined docs+tutorial website issue (unbound-force/website#350); preflight-project-checks (.uf/preflight.yaml, motivated by fullsend-ai's bash env-contract breakages) is the same shape. Riding the tutorial label on the gate's website issue satisfies both needs without issue sprawl.
