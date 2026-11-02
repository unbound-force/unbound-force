---
tag: opencode-v2-migration
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:30:21Z
identity: opencode-v2-migration-20261102T003021-jay-flowers
tier: draft
---

Migration-phase issue pattern verified while triaging unbound-force#804 (replicator Phase 3, OpenCode v2 umbrella #797): a per-sibling phase issue is operable and verifiable when it inherits the umbrella's locked decisions verbatim and pins three things. (1) Same-committed-tree compatibility: the repo must run under both binaries with one v1-shaped config (v2 auto-normalizes in memory without rewriting files; no plugins) — umbrella D1 — which prevents config forks during the transition window. (2) Version-gated dual-emission in the scaffold: the init command emits v1-shaped files for v1 runtimes and v2-native output when v2 is detected, with drift tests covering BOTH shapes so neither regresses while the window is open; this extends the tool's existing drift-detection mechanism (replicator's mirrored-file smart-sync) rather than inventing a parallel one. (3) Bounded window with clean failure ownership: the dual-emission period is scoped as a courtesy for org-internal downstream consumers (umbrella D4), not a long-term commitment, and non-plugin-surface breakage under the new binary is classified as an upstream compatibility bug to file and link back, not a local workaround. The "Done when" is testable on both axes: both binaries work on one committed tree, and init emits the correct shape per detected runtime. Sequencing constraint worth remembering: because `uf setup` installs opencode unpinned, scaffold dual-emission (Phase 2) must land before any v2 promotion — sibling phase issues inherit that ordering.
