---
tag: ci-supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-26T04:36:27Z
identity: ci-supply-chain-20261126T043627-jay-flowers
tier: draft
as_of: "unbound-force@6f63e1d4d13db641afcaf6711377a4e37823c202"
---

When reviewing a Dependabot bump of SHA-pinned reusable workflows, split the verification surface by trigger: a `pull_request`-triggered reusable workflow (e.g. deps-automerge.yml) runs from the PR merge ref for same-repo PRs, so the new pin is genuinely exercised on the bump PR itself and its passing check is real pre-merge evidence. A `workflow_dispatch`-only reusable workflow (e.g. go-release.yml) cannot run pre-merge at all, so its new contract is validated only at the next dispatch — state that surface explicitly and cover the caller-side invariants that can be checked statically (unchanged `with:` inputs, mapped secrets, job permissions, and the local guard test that asserts prefix plus 40-char SHA). Both refs in the same Dependabot group ("ship") should move together; a leftover reference at the old SHA would signal a partial upgrade.
