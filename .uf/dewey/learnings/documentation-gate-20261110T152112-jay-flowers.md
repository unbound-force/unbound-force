---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:21:12Z
identity: documentation-gate-20261110T152112-jay-flowers
tier: draft
---

A docs-gate task scoped to the current repo lets the website half of the two-repo documentation gate slip while the gate looks satisfied. Observed on merge-plugin-runtime-deps (PR #1005): task 4.3 filed only unbound-force/unbound-force#1012, and that issue deferred the website work in its body ("Remaining: website/adopter guidance..."), but the constitution's Cross-Repo Documentation rule requires the unbound-force/website issue to exist before the PR merges — a "Remaining" note in the current-repo issue is not the tracker. Authoring rule: docs-gate tasks must enumerate both gate repos (current repo and unbound-force/website), and a gate-issue body listing website coverage as "Remaining" signals the second issue still needs filing, not that the gate is met.
