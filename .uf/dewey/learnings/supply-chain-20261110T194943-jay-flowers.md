---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:49:43Z
identity: supply-chain-20261110T194943-jay-flowers
tier: draft
---

A security-override bump for a transitive exact-pinned dependency needs a three-way lockstep to be tamper-evident: (1) the version in both manifests' `overrides` blocks (canonical .opencode/package.json + scaffold mirror), (2) the provenance record (upstream git head, registry integrity hash, advisory ID, removal condition), and (3) the pinning tests — with the provenance loop comparing full structs (`got != expected`, so advisory and removal_condition are enforced verbatim), cross-checking each override version against BOTH manifests' overrides blocks, and an exact-map assertion (`assertStringMap`: length check + per-key) on the manifest overrides so extra or missing overrides fail in either direction. A bump that updates only the manifests fails provenance; one that updates only provenance fails the map equality; a silently added override fails the cardinality checks. Verified on the effect GHSA-chp3-j5qr-2vfr override (PR #1024, unbound-force).
