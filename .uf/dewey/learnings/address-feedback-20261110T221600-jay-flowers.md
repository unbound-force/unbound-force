---
tag: address-feedback
author: jay-flowers
category: gotcha
created_at: 2026-11-10T22:16:00Z
identity: address-feedback-20261110T221600-jay-flowers
tier: draft
---

When a review thread flags contradictory gatekeeping/provenance evidence, re-verify against the current PR head and base (upstream main tip) before treating the Gatekeeping "target immutable commit SHA is verified upstream" condition as unverified — an outdated thread (isOutdated: true) may describe a commit the head has already moved past. PR #1026: a reviewer flagged effect@4.0.0-beta.103 git_head ec94bd69f62e1031460ad71bb2efdf74faf6b40d in internal/scaffold/testdata/plugin-package-provenance.json contradicting tracker #1034's dff25449dfc927f2cce912c329f343cfb5365f88 (both claiming verification 2026-11-10). At head ddef2cb the shipped record (plugin-package-provenance.json:44), its scaffold_test.go twin (:1743), and the tracker all carry dff25449, and a repo-wide grep shows ec94bd69 nowhere in the tree — the contradictory value lived only on the removed reviewed commit efc07f0. Verification recipe: read the mechanism-of-record file at head, grep the tree for the flagged value, compare against the base; route residual tracker-text staleness to the tracker issue, not the PR.
