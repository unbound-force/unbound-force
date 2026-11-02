---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T00:33:54Z
identity: issue-triage-20261102T003354-jay-flowers
tier: draft
---

Triage of migration issues should verify that security-relevant motivations have matching security-relevant scope. unbound-force issue #801 (OpenCode v2 migration, Phase 2, triaged 2026-11-01) cites "internal/setup/setup.go currently installs latest opencode unpinned" as the reason dual-runtime support must land first, yet its proposed v2 install path (opt-in flag until the Phase 5 default flip) does not specify whether the new path pins a version — the unpinned supply-chain posture would silently carry into the replacement path unless pinning becomes an explicit acceptance criterion in the spec. When an issue body names an unpinned or unverified download as a problem, the fix scope should name the pinning or integrity verification that replaces it.
