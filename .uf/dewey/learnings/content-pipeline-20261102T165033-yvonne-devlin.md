---
tag: content-pipeline
author: yvonne-devlin
category: pattern
created_at: 2026-11-02T16:50:33Z
identity: content-pipeline-20261102T165033-yvonne-devlin
tier: draft
---

When a change's tasks.md scopes its documentation phase to a single issue target (e.g., uf-gate task 8.3 "File documentation issue against the current repo"), remember the documentation gate has two distinct issue requirements that should both be recorded: the in-repo docs issue (AGENTS.md documentation gate) and the unbound-force/website issue (org constitution Cross-Repo Documentation rule). In PR #582 (uf-gate), both were substantively satisfied (#583 in-repo + website#343) but tasks.md recorded only #583, leaving the website issue outside the change's audit trail. Pattern: scope documentation-phase tasks as "file documentation issues (in-repo + unbound-force/website)" and record both issue numbers in the completion note.
