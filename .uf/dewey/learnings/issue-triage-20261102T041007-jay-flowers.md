---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:07Z
identity: issue-triage-20261102T041007-jay-flowers
tier: draft
---

Cross-repo scaffold-update rollouts tracked in a meta repo triage as VALID coordination umbrellas even though no code lands in the meta repo: unbound-force/unbound-force#591 ("Update scaffolded files in complytime repositories") executes as per-repo PRs in the target organization, an established org-wide chore pattern (replicator issue #52 uf. namespace sync, website PR #169 "chore: update uf scaffolds", complytime-demos PR #80, complyctl PR #849). Triage checks that make this assessment reliable: (1) the remaining-target repository list must be enumerable — the first-triage comment supplied it; (2) related open scaffolding defects (org-content leakage #526/#527/#530, schema/config misalignment #528, stale pre-uf.* command references #589) are watch-items for the rollout, not duplicates of it; (3) the issue's "review the proposed changes and adapt what is necessary" checklist step is the security control that keeps those known defects and silent customization loss (learning scaffold-customization-20261102T024333) out of external adopter repos — per-PR review must exercise it (assert no org-specific leakage, uf.* command prefixes, schema/config alignment, deliberate KEEP/DROP of local customizations) rather than mechanically accepting uf init output.
