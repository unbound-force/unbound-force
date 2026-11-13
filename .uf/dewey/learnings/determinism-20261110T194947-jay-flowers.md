---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:49:47Z
identity: determinism-20261110T194947-jay-flowers
tier: draft
---

When a command template's shell-heredoc payload write is replaced for safety, the replacement must be an intent-level tool that owns mktemp, permissions, payload write, JSON serialization, gh api --input, and cleanup, accepting the composed text as a validated parameter — not an instructed multi-call choreography (bash mktemp/echo → host file-write → bash gh api) that threads a temp path through the agent. The unbound-force plugin suite ships the pattern three times over (file_feedback_issue, post_feedback_replies, uf_finale_create_pr); a post_issue_comment / create_child_issue tool of that shape is both the security fix and the determinism fix, and it also removes hand-JSON-escaping of untrusted text from prose. Verified on the triage-issue-heredoc change (#1011/PR #1024), where the heredoc was correctly removed but the replacement left the three-call choreography in sections 4.3/4.4.
