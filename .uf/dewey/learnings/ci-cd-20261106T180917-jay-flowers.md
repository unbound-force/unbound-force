---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:09:17Z
identity: ci-cd-20261106T180917-jay-flowers
tier: draft
---

A pinned tool version duplicated across a Makefile variable and a CI workflow step (PLUGIN_TEST_NPM in the Makefile vs the npx -y npm@<version> pin in the throwaway-lock OSV scan job) forms a twin-pin invariant. "Keep in sync" comments in both files let a one-sided bump silently desynchronize the scanned dependency tree from the tested one. Repositories that already enforce canonical-pair drift with tests (embedded assets vs canonical sources) should encode twin pins the same way: a small Go drift test in the TestPluginPackageInputs_* family (or a CI grep step) asserting the version literals match across the Makefile and the workflow file.
