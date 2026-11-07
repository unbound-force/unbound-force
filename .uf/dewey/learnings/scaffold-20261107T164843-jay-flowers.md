---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:48:43Z
identity: scaffold-20261107T164843-jay-flowers
tier: draft
---

Escape hatches added to meta-repo tooling surfaces (Makefile PLUGIN_TEST_NPM, CI workflow jobs, AGENTS.md) never reach scaffold consumers: uf init deploys only scaffold assets. A remedy a consumer needs on an activation-failure path must live in the consumer-shipped guidance layer — the scaffold/doctor remediation constants and doctor InstallHints — or it is unreachable. Two structural constraints govern edits to that layer: (1) remediation strings pinned verbatim by spec-delta scenarios (SC-FR-004's offline-install-failure scenario names the install command) must be amended in the same change that changes the string; (2) the scaffold and doctor remediation constants are twins kept in sync by a documented design commitment (D2), so a fix must move all four (scaffold runtime/install, doctor runtime/install) in lockstep — extract them to a shared internal package if they ever gain conditional logic.
