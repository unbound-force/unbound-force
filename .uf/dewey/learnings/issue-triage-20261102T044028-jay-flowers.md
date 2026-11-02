---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:40:28Z
identity: issue-triage-20261102T044028-jay-flowers
tier: draft
---

Triage of command-documentation issues requires exact command-boundary verification, not substring matching: a substring search for "uf gate" matches existing "uf gateway" documentation (a different command — LLM reverse proxy vs. constitution compliance), which can falsely suggest coverage exists. Verify the gap with a command-specific pattern (e.g., "gate --phase"), and corroborate the command's existence through its adding PR (title, summary, closes-linkage to the parent tracker) rather than the local checkout when the PR is unmerged — in-flight OpenSpec/Speckit change artifacts live on the PR branch, not main, so their local absence is expected and does not invalidate the issue. Among same-shaped open "[Docs] <command>" issues, the command identity is the duplicate discriminator; title-shape similarity alone is not duplication.
