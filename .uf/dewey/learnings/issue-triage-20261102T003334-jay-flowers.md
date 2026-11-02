---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:33:34Z
identity: issue-triage-20261102T003334-jay-flowers
tier: draft
---

Org-wide migration umbrellas (e.g., OpenCode v2 migration, umbrella #797) decompose into per-repo, per-phase tracker issues (e.g., #802 dewey Phase 3, #803 gaze Phase 3, #804 replicator, #805 website). Keyword duplicate searches surface the umbrella and sibling-phase trackers as candidates; these are structural decompositions, not duplicates. Distinguish with three checks: (1) target repo — the candidate must cover the same repository as the triaged issue; (2) parent reference — an issue whose body cites the candidate as its umbrella ("Umbrella: #797") is a child decomposition, not a duplicate; (3) phase scope — sibling trackers for other repos are parallel work with disjoint targets. Well-formed phase trackers carry Goal/Scope/Done-when sections plus explicit out-of-scope deferrals (e.g., "No native-shape conversion in this phase"), which supports a VALID verdict with enhancement category and objectively verifiable completion criteria.
