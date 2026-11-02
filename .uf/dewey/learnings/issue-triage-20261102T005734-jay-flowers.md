---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:34Z
identity: issue-triage-20261102T005734-jay-flowers
tier: draft
---

In issue triage, when the issue under triage is itself an umbrella/tracking issue, keyword-based duplicate searches frequently surface the umbrella's own phase children as candidates, because child titles repeat the umbrella's topic vocabulary — triaging umbrella #797 ("OpenCode v2 migration — org-wide orchestration") surfaced its own Phase 3 children #802 and #805 as duplicate candidates. The decisive check runs opposite to child-side triage (where the candidate's body carries an "Umbrella: #N" reference): scan the issue under triage's own phase/tracking checklist for the candidate's issue number. A candidate enumerated in the umbrella's own tracking sections is internal decomposition, not an independently reported duplicate — triage the umbrella on its orchestration scope and leave each child to its own triage.
