---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T18:50:22Z
identity: determinism-20261110T185022-jay-flowers
tier: draft
---

Removing a shell-heredoc injection vector from a command template by replacing it with mktemp + host file-write + gh api --input as agent-sequenced steps trades a security bug for a DE-005 violation: the temp-path chaining (bash echoes the path, the file-write targets it, gh api consumes it) is a judgment-free tool chain that must collapse into an intent-level tool taking the composed text as a parameter. The unbound-force repo already ships three tools of exactly this shape — file_feedback_issue, post_feedback_replies, and uf_finale_create_pr — each taking composed text as a validated parameter and performing temp-file writing, JSON serialization, gh invocation, retry, and cleanup internally. A post_issue_comment / create_child_issue tool of that shape is both the security fix and the determinism fix: it also removes the agent's hand-JSON-escaping of untrusted text, which is itself an injection-adjacent transformation when done wrong. Verified on the triage-issue-heredoc change (#1011/PR #1024), where the heredoc was correctly removed but the replacement left the three-call choreography in prose in sections 4.3 and 4.4 of uf.triage-issue.md.
