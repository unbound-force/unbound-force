---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:40:25Z
identity: issue-triage-20261102T044025-jay-flowers
tier: draft
---

In the unbound-force meta repo, open documentation issues exist for many sibling commands, so duplicate detection for a docs issue MUST match on the exact CLI command name (e.g., `uf gate`), not on generic keywords like `docs` or `gate`: the fallback search for issue #583 surfaced nine false candidates (#819 /uf.unleash fix loop, #862 /uf.fix-ci, #810 /uf.finale, #856 /uf.rebase-pr, #825 agent-brief stale gate, #908 /uf.review-pr, #747 and #838 /uf.run-pipeline-plan, #740 plugin tool parameters) where keyword overlap on "gate" or "docs" is coincidental. A docs issue is objectively triageable when it carries these anchors: a numbered spec-workflow task reference (e.g., "Closes task 8.3 from openspec/changes/uf-gate/tasks.md"), source artifact paths with the FR range (specs/gate-command.md FR-001 through FR-011), parent-issue linkage (#514), and a user-facing change (new CLI command from PR #582) that triggers the repo's documentation gate. With those anchors, the verdict reduces to checking task existence, spec coverage, and command-name-level duplicate absence.
