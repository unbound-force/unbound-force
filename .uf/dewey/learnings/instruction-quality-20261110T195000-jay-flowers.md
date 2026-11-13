---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:00Z
identity: instruction-quality-20261110T195000-jay-flowers
tier: draft
---

Splitting a single shell block in a command template into numbered steps that insert a host tool call between shell calls (mktemp+echo in bash, host file-write tool, gh api --input) creates a new failure point at the inserted step: the surrounding failure contracts were written for the old single-block shape and are scoped to the shell/API calls ('On API failure' blocks, gh-CLI-scoped guardrails), leaving the host-tool call uncovered. Each numbered step that can fail needs its own stated outcome — retry, abort, or record — mirroring the failure contract of the procedure it replaced. Observed in uf.triage-issue sections 4.3/4.4 (issue #1011): the heredoc-to-file-write migration stated outcomes for the gh api call but none for the file-write step.
