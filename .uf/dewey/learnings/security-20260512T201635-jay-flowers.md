---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-05-12T20:16:35Z
identity: security-20260512T201635-jay-flowers
tier: draft
---

In internal/sandbox/devpod.go, the startServerViaSSH function constructs a shell command string using string concatenation with the workspace name: "cd /workspaces/"+wsName. While the workspace name is sanitized by projectName() (lowercase alphanumeric + hyphens only via regex), this pattern of building shell commands via string concatenation should be flagged for review. The sanitization in workspace.go (sanitizeRe = [^a-z0-9-]) provides adequate protection against injection, but the pattern itself is fragile — any future change to projectName() that relaxes the regex could introduce a command injection vector.
