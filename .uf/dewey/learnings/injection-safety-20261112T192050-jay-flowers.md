---
tag: injection-safety
author: jay-flowers
category: gotcha
created_at: 2026-11-12T19:20:50Z
identity: injection-safety-20261112T192050-jay-flowers
tier: draft
---

Quoted heredocs ('EOF') in commands that interpolate untrusted issue bodies do not prevent delimiter breakout — payload containing a line `EOF` terminates the heredoc early and executes as shell. Fix is to extract gh payloads to host-owned 0o600 temp files via file-write tools (post_issue_comment/create_child_issue) and pass via `gh --body-file`/`--input`, scoping the security invariant to comment/child-issue creation and explicitly allowlisting sanitized non-tool channels (gh label create/edit, gh issue list --search).
