---
tag: ci-supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-26T16:35:30Z
identity: ci-supply-chain-20261126T163530-jay-flowers
tier: draft
as_of: "unbound-force@56992cd6fed3f692274a5ff3aba690f32f457308"
---

Adopting new inputs of a workflow_dispatch-only reusable release workflow has an asymmetric verification surface: the caller's guard test can only assert the literals the caller sends (e.g. `go_install`, `tap_install_check` in release.yml with `"$BIN" --version`), never that the pinned SHA actually declares those `workflow_call` inputs or exports `$BIN` in the check's context. An undeclared input fails the next dispatch at GitHub workflow validation before any tag is created, so blast radius is bounded but the one-shot pipeline still goes red; a missing shell variable fails mid-release. Record the verification of the contract against the pinned SHA in the change record, and add the re-verification of the caller's `with:` keys to the dependency-maintenance runbook so a pin bump re-checks the names. Related: learning ci-supply-chain-20261126T043627 (pre-merge verification surface for SHA-pinned reusable workflows).
