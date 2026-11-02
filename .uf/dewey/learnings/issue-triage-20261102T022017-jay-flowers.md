---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:20:17Z
identity: issue-triage-20261102T022017-jay-flowers
tier: draft
---

Documentation-gap issues that track a shipped change are most actionable when they enumerate each documentation surface explicitly — command-reference updates, the artifact contract (path scheme, schema fields, failure posture), consumer-side behavior changes, and tool-reference entries — and link both the originating spec and the cross-repo website issue. unbound-force issue #736 (persist-unleash-summary docs) exemplifies the pattern: its four-item checklist maps one-to-one onto the doc surfaces of the change and records the already-filed website issue (unbound-force/website#308), so the issue is directly executable as a single docs PR without further scoping.
