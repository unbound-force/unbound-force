---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:24Z
identity: dependencies-20261110T195024-jay-flowers
tier: draft
---

A CVE-driven dependency override that rides a feature branch (e.g., committed by /uf.fix-ci so branch CI passes OSV scanning) still needs its own CHANGELOG entry naming the GHSA id, old/new versions, verified provenance date, and rollback — the stop-shipping-plugin-lock entry is the repo template. Provenance fixtures (git_head, dist integrity) and pin-pinning scaffold tests satisfy the gatekeeping supply-chain conditions but not the documentation gate, and the hotfix spec-exemption requires exactly that retroactive documentation. When reviewing a diff that touches a package.json overrides block, grep CHANGELOG.md for the advisory id before assuming the change is documented.
