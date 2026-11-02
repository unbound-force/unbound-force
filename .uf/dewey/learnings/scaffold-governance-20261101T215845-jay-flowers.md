---
tag: scaffold-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-01T21:58:45Z
identity: scaffold-governance-20261101T215845-jay-flowers
tier: draft
---

Unbound Force model-matrix slice boundaries: the slice-1 OpenSpec change (model-matrix-delegation, issue #887) deliberately deferred hero-repo deployment to slice 2 (issue #889) and recorded design expectations that slice 2 must consciously reconcile rather than silently inherit or contradict. Two recorded tensions: (1) design decision D10 states a deployed repo's tracked .uf/model-matrix.yaml becomes "team-owned config once uf init writes it, at which point it diverges from the template by design", while #889's acceptance criterion 2 specifies tool-owned refresh semantics ("correctly refreshes tool-owned matrix files") — the re-init clobber-vs-preserve behavior for team-edited matrices must be explicitly decided inside the slice-2 change, with the #884/#886 reviewer-capabilities vs review-matrix re-run precedent as the cited prior art; (2) the meta-prompt pack (#888 deliverable) self-describes as a repo-local governance pack that "ships no -custom.md companion" (project-specific rules revised through governance change, per the determinism.md precedent), while #889 plans a meta-prompt-custom.md initial-deployment template — when the pack becomes a scaffold asset its header and companion-file policy must be updated, because hero repos cannot revise a tool-owned canonical pack and need a user-owned extension point. Triage verdicts for such slice-2 issues remain VALID (the issue's own OpenSpec change is the reconciliation venue); flag the tensions as must-decide design items rather than blockers.
