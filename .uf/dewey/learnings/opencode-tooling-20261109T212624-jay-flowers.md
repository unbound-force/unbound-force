---
tag: opencode-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-09T21:26:24Z
identity: opencode-tooling-20261109T212624-jay-flowers
tier: draft
---

Glob patterns that traverse hidden directories (e.g., `.opencode/commands/opsx-*.md`) return "No files found" even when the files exist — the glob tool excludes dot-directories from recursive matching. When verifying slash-command existence for instruction-lint IL-010 dead-reference review, read the `.opencode/commands` directory listing directly instead of globbing; trusting the empty glob on PR #1001 would have misclassified the valid /opsx-propose fix as a new dead reference.
