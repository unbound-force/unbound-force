---
tag: supply-chain
author: jay-flowers
category: decision
created_at: 2026-11-07T17:06:14Z
identity: supply-chain-20261107T170614-jay-flowers
tier: draft
---

When reviewer feedback bundles two remedies for one duplicated pin — an enforcement remedy (a drift test asserting the twins match) and a documentation remedy (record dist.integrity beside the pin) — split the acceptance instead of taking or leaving the whole item. The drift test matches the repo's executable-drift convention for constants gating two verification pipelines (PR 957: Makefile PLUGIN_TEST_NPM vs ci_security.yml osv_plugin_transitive; grep confirmed no test referenced either location) and fails loudly on one-sided bumps, so accept it. The recorded-hash remedy is content-anchored, not content-verified: nothing enforces it, it goes silently stale on pin bumps, and inside npx/npm's built-in registry verification the constitution's content-hash MUST does not bind (accepted-risk class A7) — classify that half AUTHOR-DECIDES unless it ships as an enforced check, ideally folded into the same drift test (compare `npm view npm@<version> dist.integrity` against the recorded value).
