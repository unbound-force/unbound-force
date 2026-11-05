---
tag: review-verification
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:51:19Z
identity: review-verification-20261105T225119-jay-flowers
tier: draft
---

Harvested review learnings can carry prescriptive remnants of findings that were never submitted: when a child reviewer's learning prescribes a corrective pairing (PR #955, workflow-resume-gates-20261105T214909: the descendant-acceptance re-entry gate "must be paired with push-target validation: HEAD_REMOTE must resolve to the PR's head repository"), that prescription is a finding candidate that outlives the round. Before closing a review iteration, sweep each newly harvested learning's MUST-statements against the shipped fix's gate inventory — a learning that names a missing half of a fix marks a gap no findings list carries, because child reviewers cannot submit findings after dispatch consolidation and the harvest path preserves their observation only as prose. Detection is mechanical: extract the learning's prescriptive clauses and diff them against the change's shipped checks. Precedent: PR #955 iteration 1 fixed the wrong-branch hole with the HEAD_MERGE_REF cross-check while the same round's harvested learning named the load-bearing remote check that never shipped; the gap survived to iteration 3 because only the learning recorded it.
