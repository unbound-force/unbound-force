---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:33:29Z
identity: issue-triage-20261102T003329-jay-flowers
tier: draft
---

Org-wide migrations in unbound-force decompose into per-repo phase trackers filed in the meta repo under a single umbrella (OpenCode v2 umbrella #797: Phase 3 = #802 dewey, #803 gaze, #804 replicator, #805 website, all parallel). Near-identical tracker bodies are intentional parallel decomposition, not duplication — the triage dedupe key is the target repo, not body text, and the umbrella issue is a parent, never a duplicate. Well-formed phase trackers make triage deterministic: they state Goal/Scope/Done-when, explicitly scope out later phases (e.g., "no native-shape conversion — that is Phase 5 cutover"), pre-declare the expected workflow tier (quick/opsx), and pre-route breakage handling (compat bugs filed upstream and linked back to the tracker).
