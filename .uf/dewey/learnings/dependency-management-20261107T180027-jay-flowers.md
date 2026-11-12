---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-07T18:00:27Z
identity: dependency-management-20261107T180027-jay-flowers
tier: draft
---

When a change removes a control from scaffolded output, adopters inherit the loss silently unless the compensating control is itself scaffolded or shipped as a copyable recipe. Observed on unbound-force PR #957: dropping the tracked .opencode/package-lock.json removed lock-based transitive vulnerability scanning from every hero repository uf init touches; the compensating throwaway-lock OSV job was added only to the meta repo's ci_security.yml, with distribution delegated to a CHANGELOG rollout note ("copy the osv_plugin_transitive job") and filed docs issues. Adopters who do not read the rollout note keep running with reduced CVE coverage and no diagnostic flags the gap — uf doctor checks the local install state but not the CI-control presence. Pattern: for any control removal that propagates through scaffolding, treat the compensating control's distribution path (scaffolded asset, doctor check for its absence, or at minimum a tracked migration task per adopter repo) as part of the change's definition of done, not as prose follow-up.
