---
tag: workflow-resume-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:35:59Z
identity: workflow-resume-gates-20261105T193559-jay-flowers
tier: draft
---

A command that freezes immutable context at ingest (head_sha resolved from gh pr view) and then mutates local state in a later phase (commit-then-push) deadlocks its own resume path if the re-entry gate requires exact equality with the frozen value. In PR #955, /uf.address-feedback Section 1.2.1's 'HEAD is not head_sha -> STOP' fired on the precise state a failed Section 4.4 push leaves behind (HEAD = head_sha + unpushed commits), blocking the get_progress resume before it could retry the push. Gate on ancestry instead: git merge-base --is-ancestor <frozen_sha> HEAD accepts the command's own linear progress (equal or ahead) while staying fail-closed on divergence or behind. Pin the predicate in the command contract test clause so the resume semantics cannot silently regress.
