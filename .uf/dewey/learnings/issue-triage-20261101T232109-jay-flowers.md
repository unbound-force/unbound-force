---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:21:09Z
identity: issue-triage-20261101T232109-jay-flowers
tier: draft
---

Docs-sync issue triage pattern (unbound-force #840 precedent): issues reporting stale documentation after a shipped behavior change are objectively verifiable when they quote the superseded sentences verbatim, cite the source change directory (openspec/changes/<name>/), carry a per-page documentation-gate assessment with explicit no-change/exempt verdicts, record a skip rationale for adjacent pipelines (content pipeline), and link the website counterpart for the two-target docs gate (current-repo issue + unbound-force/website issue). Related open issues editing the same doc sections for different content (e.g., #739 deterministic-tool migration vs. #840 clarify ask gate, both touching the /uf.unleash sections of docs/usage.md and docs/architecture.md) are not duplicates — the correct disposition is consolidate-don't-duplicate guidance in the issue body rather than closure as duplicate. Gotcha: quoted line anchors drift as sibling docs PRs land (issue #840 cited usage.md:152-154 and architecture.md:438-440; days later the verbatim passages sit at 173-181 and 448-455) — verify by quoted sentence content, not line number.
