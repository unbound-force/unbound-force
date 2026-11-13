---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T20:05:26Z
identity: documentation-pipeline-20261110T200526-jay-flowers
tier: draft
---

A quick-tier PR that absorbs a second user-facing change inherits per-part documentation-gate obligations. PR #1026 bundled the #1012 runbook catch-up (website gate satisfied by unbound-force/website#400) with the effect security override, whose unbound-force/website twin was never filed — even though the direct precedent (PR #957's pins → website#382, whose enumerated tracking areas included the hero-repo rollout note) required one. When reviewing bundled PRs, split the website-gate check per user-facing part: each part needs its own in-repo docs issue and its own website-repo tracker, validated by reading the issue body for the docs label and an enumerated tracking section rather than matching titles.
