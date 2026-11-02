---
tag: spec-review
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:33:37Z
identity: spec-review-20261102T033337-jay-flowers
tier: draft
---

When an OpenSpec change carries a MODIFIED-entry scenario that falsifies predecessor behavior (e.g. host-fallback-provenance's 'Host source with empty profile uses host model' superseding advisor-host-model-fallback's 'Non-advisor with empty profile rejected'), the tasks.md test mapping must pin the new scenario's exact GIVEN — including the fixture that distinguishes it (here: the empty-profile review matrix). ADDED-requirement scenarios routinely get named sub-tests, but MODIFIED-entry scenarios are easily left to an ambiguous 'mirror the predecessor block' instruction that breaks down as soon as another sub-test deviates from the predecessor's fixture convention, leaving the falsified text without a regression lock. Reviewer check: for every MODIFIED/REMOVED scenario named in the design's supersession rationale (D6-style), confirm some task specifies the fixture under which the old outcome would be observable pre-fix.
