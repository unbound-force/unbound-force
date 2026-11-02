---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:20:53Z
identity: issue-triage-20261101T222053-jay-flowers
tier: draft
---

Documentation-gate issues in unbound-force (label: docs) are the tracking artifact mandated by the AGENTS.md rule "MUST file a documentation issue against the current repo for user-facing changes before PR merge" and by the change's own tasks.md gate task; they complement the spec task list rather than duplicate it. Triage verification is objective and mechanical: (1) the named OpenSpec change directory exists; (2) each enumerated surface is greppable in shipped artifacts — new config fields across ALL schema copies (root registry schemas/, deployed .uf/schemas/, embedded internal/scaffold/assets/), tool input/output contracts in the plugin source and its byte-synced scaffold mirror, command prose in .opencode/commands/; (3) the CHANGELOG entry exists where the issue says it does, and a well-filed gate issue is cross-referenced from that entry by issue number; (4) the target documentation file (AGENTS.md) does not yet mention the surfaces, proving the tracked work is pending rather than already complete. Surface bullets carrying FR/DR identifiers give the documentation task a mechanically checkable acceptance list. No split is warranted when all surfaces belong to one change and converge on one documentation file. Verified on issue #859 (derive-per-run-timeout): all five surface bullets checked out, including the exact three-schema-copies count and the CHANGELOG cross-reference.
