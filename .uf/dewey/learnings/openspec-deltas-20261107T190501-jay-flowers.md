---
tag: openspec-deltas
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:05:01Z
identity: openspec-deltas-20261107T190501-jay-flowers
tier: draft
---

Spec-delta scenario clauses that name an escape hatch should pin the invariant — the existence of a pinned-runner escape hatch for the named failure mode — rather than the runner's exact version literal. Precedent: stop-shipping-plugin-lock's amended SC-FR-004 offline-install-failure scenario (openspec/changes/stop-shipping-plugin-lock/specs/scaffold/spec.md) now reads 'remediation offers npx -y npm@11.6.2', the same literal the scaffold/doctor constants and their parity test carry. The constants and tests move together mechanically on a bump (TestRemediationConstants_CarryNpm11EscapeHatch fails until all four constants update), but nothing couples the archived scenario text to them: after archive, a security-driven pin bump leaves the capability spec asserting a stale version unless a new change amends it. This instantiates the existing openspec-deltas rule for the escape-hatch clause shape specifically: the version belongs in the constants and their guard, the scenario names the runner class ('the pinned npm 11 runner') plus the failure mode it exists to bypass.
