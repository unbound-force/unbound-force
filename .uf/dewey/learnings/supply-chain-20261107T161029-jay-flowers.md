---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:10:29Z
identity: supply-chain-20261107T161029-jay-flowers
tier: draft
---

When a lockfile is removed in favor of lockless installs, the supply-chain posture holds only if every install path independently carries the hardening flags and the compensating scanner resolves the same tree the installs produce. In unbound-force's stop-shipping-plugin-lock change: (1) --ignore-scripts had to be added to each of the three install surfaces separately — uf init activation, make plugin-test, and the CI throwaway-lock generation — because no shared config carries it; (2) the throwaway-lock OSV scan must run with the same pinned npm version as the installs it compensates for (npx -y npm@11.6.2 in both Makefile PLUGIN_TEST_NPM and the workflow), since npm resolution differs across majors and a desynced pin would scan a different transitive tree than the one shipped. Twin pins spanning files need a drift test, not just "keep in sync" comments.
