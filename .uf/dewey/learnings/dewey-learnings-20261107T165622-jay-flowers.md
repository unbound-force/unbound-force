---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:56:22Z
identity: dewey-learnings-20261107T165622-jay-flowers
tier: draft
---

PR 957 shipped retrospective learning scaffold-testing-20261106T200305 — which prescribes pairing preservePluginDevDependencies's unit tests with one orchestrator-level Run() test that pre-populates an existing manifest declaring devDependencies and asserts the deployed bytes plus Created/Updated/skip reporting — as a learning file inside the same PR that omitted the prescribed test; the gap surfaced only in post-PR feedback triage. A retrospective learning that names a concrete missing test, gate, or code remedy is knowledge capture, not verification. Trace the prescribed remedy to a tasks.md item in the same change before the PR ships, or record in the learning why it is deferred; otherwise the change knowingly ships the gap its own lesson documents.
