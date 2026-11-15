---
tag: drift-tests
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:24:24Z
identity: drift-tests-20261110T212424-jay-flowers
tier: draft
---

Pin-sweep drift tests pass vacuously when the swept surface drops the pinned item entirely: a loop over regex matches enforces parity only for pins that are present, so pin REMOVAL from the swept file stays green. Every parity sweep needs a paired presence assertion (fail when zero matches are found). Likewise, when a sweep iterates a hand-maintained list of guarded surfaces, parity among the listed files does not close the failure class — only a negative-space test does (walk the enclosing directory, e.g. .github/workflows/*.yml, and fail when the guarded markers appear outside the list). Observed in unbound-force PR #1006 (plugin-lock-advisories): the new sweeps enforce presence for both CI workflow copies but not for the copyable recipe doc, and the hardcoded osvTransitiveWorkflows list means a future fourth copy of the throwaway-lock scan would again ship unswept — the exact gap (issue #991 #1) the change was created to close.
