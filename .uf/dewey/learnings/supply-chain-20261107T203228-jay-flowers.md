---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:32:28Z
identity: supply-chain-20261107T203228-jay-flowers
tier: draft
---

A retroactive supersession change that takes its sweep scope from the triggering review thread's findings inherits the thread's blind spots: review threads name behavior contracts, not pin records. In PR 948, retroactive-push-skill-contracts annotated exactly the five deltas the round-2 spec-consistency finding named, while the same changeset's zod 4.1.8->4.6.6 bump and rollback-baseline re-point left clarification-marker-superset SC-FR-003 — a dependency-pin carrier no thread named — pinning the superseded values. Derive the sweep enumeration from the values the changeset moves, not from the thread's delta list: grep the unarchived openspec/changes/ stack for each old version string and whole-file hash the diff touches; the moved-values set is the complete carrier set, the thread-named set is only the noticed subset. Scope-claim text in the retroactive record ("five live unarchived deltas") should then state a grep-derived count, so the record's own audit trail does not launder an incomplete enumeration into an authoritative one.
