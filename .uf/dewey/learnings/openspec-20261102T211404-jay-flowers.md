---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-02T21:14:04Z
identity: openspec-20261102T211404-jay-flowers
tier: draft
---

When an OpenSpec change restates a MODIFIED requirement, enumerate ALL unarchived deltas for that capability — not just the origin and extend changes — before writing the restatement. Each unarchived MODIFIED delta carries a full-requirement replacement the new change must compose with; a restatement built from stale text silently reverts the sibling change's clauses at archive time (OpenSpec replaces requirements wholesale; the last-archived restatement wins). Seen in fix-uf-finale-ci-poll-gh-schema: its MODIFIED FR-006 resurrected the pre-RD-FR-032 exceeds_threshold wording that the unarchived review-dispatch-diff-manifest delta had explicitly repealed, creating a present-tense contradiction between two unarchived deltas and a deterministic clause-reversion path regardless of archive order.
