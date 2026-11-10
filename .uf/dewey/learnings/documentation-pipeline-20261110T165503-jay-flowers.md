---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T16:55:03Z
identity: documentation-pipeline-20261110T165503-jay-flowers
tier: draft
---

unbound-force CHANGELOG entries follow a two-repo Docs citation convention: "(Docs: #<in-repo issue>, unbound-force/website#<n>)" — 20+ precedent entries cite both the repo-local docs-gate issue and the website tracking issue. Filing both issues satisfies the documentation gate, but the CHANGELOG entry should cite both too; otherwise the website tracker is discoverable only from the change's tasks.md. Verified on PR #1007: the entry cites "(Docs: #1003)" alone although unbound-force/website#401 exists (open, docs label, Pages-to-update section) and tasks.md task 4.2 records both.
