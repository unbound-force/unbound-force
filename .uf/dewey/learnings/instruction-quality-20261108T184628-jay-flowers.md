---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:46:28Z
identity: instruction-quality-20261108T184628-jay-flowers
tier: draft
---

Prescribed GitHub API mutations in command prose must not compose JSON payloads by raw heredoc/textual substitution of finding or review text: double quotes, backslashes, and newlines in that text produce invalid JSON and a 4xx from the API. Compose payloads with an escaping-aware mechanism (jq -n --arg, gh api -f form fields, or gh issue create --title/--body-file) while keeping the temp-file + --input shell-injection discipline — a quoted heredoc prevents shell injection but does not make the payload valid JSON. Additionally, whenever a dependent artifact is only legal after the mutation succeeds (e.g., a FILE reply that must link a just-created issue), the procedure must state the failure outcome: retry once; on second failure, do not post the dependent artifact — stop and report. Precedent: the uf.address-feedback FILE issue-filing block feedback item (PR #993, HIGH), where the same command's reply-posting block already paired temp-file + --input with a crash-recovery paragraph but the new issue block omitted both the escaping and the failure branch.
