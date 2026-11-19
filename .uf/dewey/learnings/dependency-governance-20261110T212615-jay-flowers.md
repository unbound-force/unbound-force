---
tag: dependency-governance
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:26:15Z
identity: dependency-governance-20261110T212615-jay-flowers
tier: draft
---

A provenance-record-to-manifest join is one-directional by default: iterating provenance entries and asserting each is declared in the manifest catches stale provenance records, but not ungoverned manifest entries. Close the lattice in both directions — a bidirectional map assert (length equality plus per-key comparison) on the manifest's override block, plus a length-equality check on the provenance record — so a dependency override can only be added by editing the gate tests themselves, which is a review-visible change under the Gatekeeping rule.
