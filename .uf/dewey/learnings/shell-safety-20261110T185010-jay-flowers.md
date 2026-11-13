---
tag: shell-safety
author: jay-flowers
category: pattern
created_at: 2026-11-10T18:50:10Z
identity: shell-safety-20261110T185010-jay-flowers
tier: draft
---

A quoted shell heredoc is not injection-safe for composed payloads: a heredoc terminates at any line equal to its delimiter even when the delimiter is quoted, so payload text containing that line breaks out and executes as shell. Remediation pattern (uf.triage-issue #1011): create the temp file with mktemp + chmod 600 in one shell call and echo the path (shell variables do not persist between separate shell calls); write the JSON payload with the host's file-write tool; post via gh api --input with the printed path quoted; rm -f in all exit paths. Enforce with a contract test that regex-rejects heredoc operators across command templates, and keep untrusted data out of script-layer heredocs (council-review-action build-prompt.sh injects the PR title via printf '%s' outside its static heredocs).
