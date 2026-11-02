---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:34:41Z
identity: issue-triage-20261102T013441-jay-flowers
tier: draft
---

Cross-repo extension of the tutorial-vs-docs duplicate pattern (extends learnings issue-triage-20261101T221541 and issue-triage-20261101T225744): triaging a meta-repo tutorial issue can surface a website-repo docs issue via prior-knowledge search that enumerates the same workflow steps. Verified triaging unbound-force/unbound-force#767 (tutorial: recovering a stranded review dispatch): unbound-force/website#313 (docs: review-dispatch recovery tools) lists the same recovery steps — correlation_id resolution from a fresh parent session, discard_orphaned_submissions vs redispatch_failed_runs, consolidate_dispatch warnings. The pair remains complementary (public website reference vs. in-repo operator tutorial; different repos, audiences, and artifact homes), but a docs issue whose body enumerates workflow steps narrows the deliverable gap. Triage should flag the cross-repo pair for explicit scope coordination before writing begins and record which repo canonically carries which deliverable, so the two efforts do not produce overlapping prose.
