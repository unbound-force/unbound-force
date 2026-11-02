---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-01T22:19:47Z
identity: issue-triage-20261101T221947-jay-flowers
tier: draft
---

Quoted repository-state counts in documentation-gate disposition comments are point-in-time evidence, not current state — triage must date-reconcile them before flagging drift. Verified while triaging unbound-force issue #873 (2026-11-01): the task-5.3 comment (2026-10-29) quotes AGENTS.md's structure map as "Hero persona agents (16 active)", but HEAD's AGENTS.md reads "17 active, model-matrix-classifier hidden" — the delta is the subsequently-landed model-matrix-delegation change adding the hidden classifier agent, not an error in the comment. Recipe: when a disposition comment's quoted count disagrees with current repo state, check whether a later change (higher issue number, later CHANGELOG Unreleased/Added entry) explains the drift before questioning the disposition; flag a finding only when the count was wrong at the comment's own date. Same discipline applies to structure-map counts the comment itself flags as pre-existing (e.g. AGENTS.md "Slash commands (33 files)" vs 34 files on disk) — record as out of scope, as #873's comment correctly did, rather than attributing it to the tracked change.
