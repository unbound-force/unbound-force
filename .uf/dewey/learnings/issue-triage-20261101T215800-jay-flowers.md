---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T21:58:00Z
identity: issue-triage-20261101T215800-jay-flowers
tier: draft
---

In the unbound-force meta repo, docs-labeled issues filed for user-facing behavior changes are documentation-gate artifacts with two independent mandates: the AGENTS.md behavioral rule (file a documentation issue against the current repo before PR merge) and an explicit documentation task inside the originating change's tasks.md that carries the full filing recipe — unfiltered open-issue keyword search (label-filtered searches miss unlabeled docs issues), disposition of related issues, gh issue create with the docs label, and a body enumerating the user-visible changes. Triage recipe for such issues: match title/label/body against the referenced openspec/changes/<name>/tasks.md documentation task, confirm the spec directory exists on disk, and confirm the duplicate search was unfiltered. A match triages as VALID/enhancement/objective. Cross-reference notes asking a sibling docs issue (e.g., a recovery-surface issue covering redispatch_failed_runs and consolidate_dispatch) to mention the change are coordination signals, not scope creep, and do not justify a split.
