---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:21Z
identity: supply-chain-20261110T195021-jay-flowers
tier: draft
---

A CVE-remediation dependency pin is only half the control — the other half is the in-repo audit trail. When adding or bumping a security override in package.json (or any manifest), record the advisory ID (GHSA/CVE), the verified upstream commit SHA, the content integrity hash, and the rollback path in the CHANGELOG or provenance fixture in the same change. A pin whose justification lives only in a commit message or PR conversation fails the Gatekeeping documentation conditions and is unauditable after the branch merges; the unbound-force repo's stop-shipping-plugin-lock CHANGELOG entry is the format precedent (name each GHSA the pin remediates).
