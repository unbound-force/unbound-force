---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:33:25Z
identity: scaffold-20261107T173325-jay-flowers
tier: draft
---

Compensating controls implemented in repo-local CI do not propagate through scaffolds. uf init deploys manifests, plugin sources, commands, and agents — never GitHub workflow jobs — so when a change removes a shipped artifact that hero repos relied on (the tracked .opencode/package-lock.json that gave OSV-Scanner and Dependabot transitive dependency scanning), the repo-local compensating control (the throwaway-lock osv scan in ci_security.yml) must be paired with a copyable recipe in the rollout guidance, or hero repos following the rollout inherit the security trade-off with no remedy path. When a CI job is offered as a copyable recipe, adapt repo-local assumptions first: go-version-file: 'go.mod' presumes a Go repo, Makefile pin-twin comments do not exist downstream, and SHA-pinned action references must stay SHA-pinned in the copy.
