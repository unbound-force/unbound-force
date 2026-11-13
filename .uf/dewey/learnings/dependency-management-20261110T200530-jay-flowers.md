---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:05:30Z
identity: dependency-management-20261110T200530-jay-flowers
tier: draft
---

In unbound-force, the runbook's override-verification command in docs/osv-transitive-scan.md is the only override-verification path — uf doctor's manifest-install check reads direct pins and never `overrides` — yet no gate covers it: when the effect 4.0.0-beta.103 override landed (PR #1026), the `npm ls zod toml` command, its expected-version sentence, and the section heading all went un-swept while every test-enforced surface (both package.json manifests, the provenance JSON, the scaffold expectedOverrides maps, CHANGELOG) was updated. Because the overrides set is machine-readable in .opencode/package.json, this doc enumeration can be gate-checked instead of grepped: a scaffold drift test can parse the runbook's fenced `npm ls` line and assert it names every key in the manifest's `overrides` block, turning the manual sweep into a deterministic check for future override additions.
