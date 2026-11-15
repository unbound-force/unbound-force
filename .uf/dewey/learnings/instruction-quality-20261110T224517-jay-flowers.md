---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:45:17Z
identity: instruction-quality-20261110T224517-jay-flowers
tier: draft
---

When a command template's payload write is migrated from a shell heredoc to the host's file-write tool (mktemp + echo path in one shell call, file-write to the printed path, gh api --input with the path quoted), the new write dependency needs its own stated failure outcome and cleanup must be reachable from that failure path. Placing `rm -f` only inside the posting bash block strands the temp file when the write never completes, and an agent left holding composed untrusted text with an empty temp file may improvise a shell write — the exact vector the heredoc removal closed. Sound pattern, one sentence per site: "If the write fails or is denied, do not fall back to a shell write and do not post: report the failure, record the action as not executed in `actions_taken`, and `rm -f` the temp file." Observed on PR #1024 (uf.triage-issue.md sections 4.3 and 4.4), extending instruction-quality-20261108T184634 (outward-facing mutation steps need explicit failure outcomes) to the write-dependency and cleanup-reachability specifics of the heredoc-to-file-write migration.
