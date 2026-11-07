---
tag: learning-capture
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:10:02Z
identity: learning-capture-20261107T171002-jay-flowers
tier: draft
---

Bulk learning captures made through 2026-11-05T14:45 truncated the learning body at ~512 characters, silently clipping the final clause of every learning that exceeded the cutoff — 20 of 44 artifacts in unbound-force PR #948 (batches 2026-11-03T21:31, 2026-11-04T20:08, 2026-11-05T14:45), with the Dewey index serving the same clipped text into retrieval. Diagnostic signature: every clipped body ends mid-token at ~508-513 characters, and the only complete 14:45 capture was complete solely because its body fell under the cutoff. Guard: after any bulk learning capture, run a tail-integrity loop over .uf/dewey/learnings/*.md flagging files whose final line lacks terminal punctuation, before committing. Remediation precedence when restores are needed: the dispatch-artifact store of the workspace that ran the review rounds (a fork checkout's store may hold only later, unrelated artifacts) > the GitHub review threads where the underlying findings were posted > re-deriving the final clause from the learning's cited PR/issue anchors; after restoring files, reload the Dewey index so semantic retrieval stops serving the clipped text.
