---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-03T11:11:04Z
identity: documentation-pipeline-20261103T111104-jay-flowers
tier: draft
---

The pre-merge documentation gate is verifiable from the PR diff alone when an OpenSpec change carries a file-the-docs-issue task: check the CHANGELOG entry for docs-issue cross-references (the unbound-force repo convention cites "Docs: #NNN" alongside "Fixes: #NNN") and the tasks.md checkbox state. On unbound-force PR #929 (bound-consolidate-dispatch-output), the CHANGELOG entry cited only "Fixes: #922" with no docs-issue references, and task 4.3 (file the in-repo + unbound-force/website docs issues) stayed unchecked — evidence the documentation gate was unmet at review time.
