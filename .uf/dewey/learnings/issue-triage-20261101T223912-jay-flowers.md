---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:39:12Z
identity: issue-triage-20261101T223912-jay-flowers
tier: draft
---

A documentation-impact issue's validity is efficiently established through a three-link verification chain: (1) provenance — the originating change's tasks.md documentation phase names the issue as a mandated deliverable (uf-fix-ci task 3.4: "File both documentation issues required by the stacked documentation gates"); (2) disposition parity — the issue's per-file dispositions (CHANGELOG.md, AGENTS.md, docs/usage.md, docs/architecture.md, QUICKSTART.md, README.md) mirror that task list and are corroborated by current repo state (the AGENTS.md PR Review Commands row and the .opencode/commands file count); (3) reasoned content-coverage declines — tutorial/blog dispositions judged against the capability-family baseline (per learning content-pipeline-20261016T172845-jay-flowers) are valid when explicit with rationale, even for a multi-phase command workflow, when the command complements an existing family rather than debuting a new capability. When all three links hold, a website tracking issue is referenced, and the duplicate search is clean, the verdict is VALID on objective evidence and the issue is closable alongside the implementing PR.
