---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-05T14:45:25Z
identity: openspec-20261105T144525-jay-flowers
tier: draft
---

A quick-tier fix that changes behavior specified in a pending (completed-but-unarchived) openspec delta must reconcile the delta text in the same change, or the main spec lands MUST text contradicting the shipped implementation at archive time. The mechanism is superseded-by annotations or an amending delta — precedent: openspec/changes/unleash-deterministic-tools tasks 9.2/9.3, which annotate stale sibling delta text precisely so "the main spec never lands text contradicting the merged contract". Seen in P
