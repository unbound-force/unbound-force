---
tag: change-records
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:18Z
identity: change-records-20261108T213318-jay-flowers
tier: draft
---

A review-fix round that adds a new lint rule (or any user-visible gate behavior) to a tool the change ships must extend the change record in the same commit: proposal What Changes, tasks.md with the round marker, spec delta if requirement-level, and the CHANGELOG entry. On PR #959's round 2, every other hardening addition (--changed scoping, IL-050 anchoring, provenance stamps, zod validation, cat-file header validation) carries a round-2 record, but the new error-severity rule IL-043 (unclosed-fence detection) shipped in none of them — an unrecorded error-level rule is invisible to future change-record audits and surprises adopters running the linter directly. The retrospective record's value is bounding what shipped; a fix round that grows the shipped surface must grow the record with it.
