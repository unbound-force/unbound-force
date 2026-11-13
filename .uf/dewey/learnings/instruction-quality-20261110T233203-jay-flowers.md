---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:32:03Z
identity: instruction-quality-20261110T233203-jay-flowers
tier: draft
---

Splitting a single shell block into a multi-call choreography (shell mktemp → host file-write tool → shell gh api --input) to eliminate a heredoc creates a new failure surface at each seam, but the file's existing failure rules were written for the old single-block shape: rules scoped to "API failure" or "gh CLI failure", artifact-contract causes enumerating only "user abort or API call failed", and cleanup comments anchored in the final block all fail to cover the new intermediate host-tool write. Any prose rewrite that separates previously-atomic steps must add a failure outcome for each new seam — stop/report semantics, actions_taken cause coverage, and cleanup reach for aborts that never reach the final block — or route the whole choreography through a tool that owns failure-stop and cleanup. Observed reviewing PR #1024 (uf.triage-issue.md 4.3/4.4): the heredoc removal was correct, but step 2's file-write failure has no stated outcome anywhere in the load set.
