---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-11-10T22:16:08Z
identity: address-feedback-20261110T221608-jay-flowers
tier: draft
---

An isOutdated feedback thread whose referenced file is absent from the current PR diff can allege a state that existed only on a since-removed commit. On PR #1026, a curator finding cited provenance git_head ec94bd69f62e1031460ad71bb2efdf74faf6b40d from internal/scaffold/testdata/plugin-package-provenance.json (review commit efc07f0), while the current head and base carry dff25449dfc927f2cce912c329f343cfb5365f88 — matching tracker #1034 — and a repo-wide grep finds ec94bd69 nowhere in the tree. Before treating such an alleged contradiction as live, read the referenced artifact at the current head and grep the alleged value repo-wide; when the value exists nowhere, the reconciliation already happened in the PR's history and only tracker-text staleness remains to fix.
