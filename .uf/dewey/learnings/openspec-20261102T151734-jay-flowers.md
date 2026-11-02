---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T15:17:34Z
identity: openspec-20261102T151734-jay-flowers
tier: draft
---

When an OpenSpec change MODIFIES a requirement whose lineage never carried an FR identifier, the full replacement text silently perpetuates the gap: OpenSpec tooling matches MODIFIED entries by requirement name, so nothing fails, and each successive change tends to reproduce the predecessor's shape. Verified in the split-host-resolution-failure-classes delta (2026-11): the Advisor Host-Model Fallback requirement has been unnumbered since advisor-host-model-fallback (#693) and stayed unnumbered through two MODIFIED entries, while sibling requirements in the same deltas carried `**ID:**` lines per MP-013. When reviewing or authoring a delta, check that every MODIFIED entry opens with an id line; for genuinely unnumbered requirements, record the unnumbered status explicitly in the MP-014 allocation header (the `dispatch_agent_run Tool` note is the in-repo precedent) so the numbering map stays complete and reviewers can distinguish 'never numbered' from 'number dropped'.
