---
tag: constitution-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:21:38Z
identity: constitution-governance-20261102T052138-jay-flowers
tier: draft
---

Composability claims in unbound-force operate at two layers that are easy to conflate during triage: heroes (repo-level tools — Gaze, Dewey, Replicator, uf — each a separately installable binary) and personas (agent files deployed by `uf init` — Cobalt-Crush, Mx-F-coach, constitution-check). Constitution Principle II v1.4.0 guarantees independent installability for *heroes* ("Every hero MUST be independently installable"), which is satisfied by the separate binaries; the *agent*-level wording the composability audit in issue #559 quotes comes from VISION.md, and persona-level install granularity genuinely does not exist beyond the `--divisor` carve-out. When triaging or reviewing composability-gap claims, first identify which layer the quoted text targets before classifying a claim as overstated. Related drift hazard: long-lived issues cite line numbers and flag counts that go stale as the codebase evolves (#559 cited scaffold.go L478/L1422 — now L785/L1881 — and three `uf init` flags; there are now five, though `--stealth`/`--check` are not persona-selection flags). Re-verify substance, not coordinates.
