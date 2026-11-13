---
tag: command-templates
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:45:12Z
identity: command-templates-20261110T224512-jay-flowers
tier: draft
---

When remediating the same vulnerability class across sibling command templates, match the remediation shape already shipped for that class. In unbound-force, #1007 removed the uf.address-feedback.md heredoc sites by routing all mutations through the intent-level tools file_feedback_issue and post_feedback_replies (composed text as parameters, tool-owned temp files, JSON encoding, and cleanup), while #1024 rebuilt uf.triage-issue.md 4.3/4.4 as instructed multi-call choreography (bash mktemp/echo -> host file-write -> bash gh api). Two shapes for one vulnerability class split the mechanism of record and leave the prose choreography as the pattern future templates copy. Before choosing prose choreography for a security remediation, check how the sibling remediation for the same class was shaped.
