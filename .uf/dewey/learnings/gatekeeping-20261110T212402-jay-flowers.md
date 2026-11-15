---
tag: gatekeeping
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:24:02Z
identity: gatekeeping-20261110T212402-jay-flowers
tier: draft
---

A security-override pin committed onto an active OpenSpec change branch falsifies that change's own scope statements unless the artifacts move with the commit. Blanket negatives in proposal Impact sections and design Gatekeeping notes ("No CI workflow, pin, threshold, or flag is changed") are audit-citable claims about the delivered branch, and a riding pin commit turns them false. The repo's dual-mandate rule compounds this: pinned values are mandated by BOTH tests and the unarchived spec text (e.g. SC-FR-003's override enumeration "The toml, zod, and uuid security overrides MUST remain"), so a new override must also amend that enumeration in the same PR — otherwise the pin ships with a test mandate but no spec mandate, and post-archive nothing in the baseline spec requires it to remain. Rule: when a Gatekeeping-exception pin rides a change branch, sweep (1) the host change's proposal Impact and design Gatekeeping claims, and (2) the override/pin enumerations in the unarchived delta that mandates the set, alongside the usual manifests, mirrors, provenance, tests, CHANGELOG, and docs.
