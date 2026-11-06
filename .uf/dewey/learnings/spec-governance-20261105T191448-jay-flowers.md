---
tag: spec-governance
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:14:48Z
identity: spec-governance-20261105T191448-jay-flowers
tier: draft
---

Before writing a cross-change reconciliation in an OpenSpec change, sweep unarchived sibling deltas for three term classes: the artifact filename (package-lock.json), the replaced command (npm ci), and renamed identifiers (manifest-lock). Classify each hit as forward-contract text (reconcile in the same PR) or historical record (leave unchanged). A sibling delta naming the changed artifact is a hidden fifth party to the reconciliation.
