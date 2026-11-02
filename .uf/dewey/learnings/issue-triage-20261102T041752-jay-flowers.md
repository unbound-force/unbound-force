---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:17:52Z
identity: issue-triage-20261102T041752-jay-flowers
tier: draft
---

Keyword-search duplicate checks miss same-scope follow-up issues: `gh issue list --search` ranks by relevance and returns 30 results per page, so an older tracker can stay invisible while a newer near-identical one is the only match surfaced. In unbound-force/unbound-force, #584 (filed 2026-09-08 for parent story #514's deferred uf-gate wiring+docs acceptance criteria) and #869 (filed 2026-10-29, same two ACs plus a multi-change-resolution decision and an optional CI coverage-ratchet note) both survived independent duplicate searches — the caller's zero-result keyword check and a prior triage session of #869 each missed the pair. A single targeted `gh issue view` of an issue number cited in a prior Dewey learning exposed the duplication. During issue triage, treat every issue number cited in the issue body, in referenced PRs, and in prior learnings as a mandatory duplicate probe, because deferred-acceptance-criteria trackers cluster around the same parent-story issue numbers, and relevance-ranked search cannot be relied on to surface them.
