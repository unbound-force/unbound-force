---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:34:52Z
identity: issue-triage-20261102T013452-jay-flowers
tier: draft
---

Documentation requests for a shipped change must be triaged against every open documentation tracker for that change across BOTH the code repo and the website repo, not just same-repo keyword matches. Observed triaging unbound-force/unbound-force#767 (tutorial: stranded-dispatch recovery): the tutorial's five sections mapped completely onto existing open trackers — the same-repo docs issue #766 ("docs: review-dispatch recovery tools and consolidation changes") and the website repo's user-facing docs issue unbound-force/website#313, whose stated scope matched the outline near-verbatim (correlation_id from a fresh parent session, choosing discard_orphaned_submissions vs redispatch_failed_runs, consolidate_dispatch re-run with orphaned_submissions_excluded warnings) — with partial overlap in #808 (correlation-lifecycle docs). Procedure: enumerate the underlying change's open docs issues across repos, map each section of the requested document against candidate scopes, and when every section maps, classify as duplicate and fold the genre requirements (audience, narrative structure, prerequisites) into the surviving tracker. Genre differences (tutorial vs reference) are presentation choices within one documentation deliverable, not new content scope; zero-waste mandates favor one consolidated tracker per deliverable.
