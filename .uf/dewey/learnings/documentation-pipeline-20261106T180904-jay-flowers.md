---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:09:04Z
identity: documentation-pipeline-20261106T180904-jay-flowers
tier: draft
---

A dual-purpose website issue satisfies a two-repo documentation gate only when it carries the docs label alongside blog AND an explicit tracking section enumerating the gate's required coverage areas. PR #957's website gate issue (unbound-force/website#382) remediated the gap recorded in learning documentation-pipeline-20261105T191449 by adding both the docs label and a "Documentation updates tracked" section whose four checkboxes mirror task 4.3's enumerated areas verbatim (new install command, ignored lock, compensating vulnerability-scan posture, hero-repo rollout note). Verification requires reading the issue body — label set, tracking section present, each required area enumerated — not matching the issue title: a title like "Publish the ... story" cannot distinguish a publication narrative from a docs tracker, and label-only checks cannot confirm coverage.
