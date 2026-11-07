---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:14:27Z
identity: supply-chain-20261107T191427-jay-flowers
tier: draft
---

A changeset that ships both a dependency-pin bump and a retroactive contract-record change should sweep the unarchived spec stack for every carrier of the values it moves — dependency versions, whole-file hashes, and record-semantics contracts — not just the carriers the triggering review thread named. Executable carriers (test consts, testdata, provenance, README/AGENTS lines) go green in lockstep, so the only residual signal is an unarchived delta's MUST text, which surfaces only at archive time as a materialized requirement contradicted by the shipped tree. Instance: PR 948's zod 4.1.8 → 4.6.6 bump updated the test const, testdata, provenance, and AGENTS.md but left clarification-marker-superset's SC-FR-003 MUST text pinning Zod 4.1.8 and rollback baseline 5ebee4af… against the shipped 4.6.6/1bdb0793…, even though the same changeset's retroactive-push-skill-contracts change annotated five other contradicted deltas, and its own supply-chain learning recorded the omission. Remedy pattern: when moving a pin, grep the unarchived openspec/changes/ stack for the old value (version string and hash) and extend the superseded-by annotation or a MODIFIED pin-record delta to every hit in the same change; when a hash record's semantics must change (rollback-target → current-state), record the semantics change beside the re-pointed digest rather than redefining it silently in a test comment.
