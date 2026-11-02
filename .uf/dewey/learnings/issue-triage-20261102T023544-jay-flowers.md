---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:35:44Z
identity: issue-triage-20261102T023544-jay-flowers
tier: draft
---

Docs-tracking issues that reference different merged PRs can touch the same documentation files without being duplicates. In triage of unbound-force/unbound-force #706 (document checkpoint/get_progress/clear_progress tools from PR #705), duplicate candidates #739 (deterministic-tool migration docs, PR #716 Increment 2) and #740 (restored tool parameter behavior) also modify uf.unleash/uf.finale documentation, but each documents a distinct shipped change. Duplicate detection for documentation issues must compare the underlying change being documented (the referenced PR/change record), not the files touched.
