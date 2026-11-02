---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:38:38Z
identity: issue-triage-20261101T233838-jay-flowers
tier: draft
---

Docs-gate issues in unbound-force are often the literal deliverable of a documentation-gate task inside the originating OpenSpec change's tasks.md (example: issue #827 is fix-init-speckit-sentinel task 4.5(a), marked [x] at filing time). Triaging them is mechanical when the issue body is cross-checked against that task: the task enumerates the required pages, the exact behavior deltas, the filing label, and the required disposition of related open issues. Three greps verify validity: (1) the change directory exists under openspec/changes/<name>/; (2) the cited stale wording is still present at the named doc location (work pending) or already replaced (work done, issue closeable); (3) the config key or flag the issue asks to document appears in code/tests and the change's artifacts but not yet in docs/. The task's checkbox state distinguishes procedural gate filing (checked) from pending gate work (unchecked). Watch for the paired website-repo gate task (org constitution Cross-Repo Documentation rule) that files a separate tracking issue in unbound-force/website.
