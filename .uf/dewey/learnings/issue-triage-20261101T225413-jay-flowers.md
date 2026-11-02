---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:54:13Z
identity: issue-triage-20261101T225413-jay-flowers
tier: draft
---

Documentation-gate tracking issues are objectively triageable when they carry three structural elements: (1) traceability to the spawning change — task id, branch, and artifact path, with the change's tasks.md recording the filed issue number so the traceability loop closes in both directions; (2) an exhaustive per-surface disposition table covering every candidate docs surface (CHANGELOG, each docs/ file, AGENTS.md/README.md, content coverage, website) including explicit no-change entries with rationale; and (3) acceptance checkboxes tied to verifiable artifacts in the implementing PR — a merged CHANGELOG entry, an updated file, or a recorded no-change disposition. unbound-force/unbound-force#844 (replicator-version-floor docs impact) exhibits all three, which let a triager confirm VALID purely against repo evidence (change directory, tasks.md state, CHANGELOG entry, docs file contents) without subjective judgment.
