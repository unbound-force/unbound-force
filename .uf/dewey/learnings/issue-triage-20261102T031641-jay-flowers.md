---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:16:41Z
identity: issue-triage-20261102T031641-jay-flowers
tier: draft
---

Plugin-tool extraction issues can be partially preempted by intermediate migrations and still remain valid — with shifted rule relevance. unbound-force #668 (resolve_conflict) cited uf.finale.md lines 625-953; the later finale-token-reduction change shrank the file to 399 lines and replaced raw git prose with granular tool chaining in section 7b (`git_fetch → git_merge; if conflicted → git_merge_abort`). The issue stays VALID but escalates from DE-001 (prose→tool) to DE-005 (judgment-free tool chain→intent-level tool) relevance: the current per-option sequences are fully determined by prior structured results, exactly what DE-005 MUST-collapse sanctions, and the proposed intent-level tool's strategy parameter is the DE-006 judgment-product boundary (option selection stays at the human gate). Triage checks for preempted extraction issues: (1) grep the current command for tool-arrow chains where the cited prose was — citation drift with verifiable content is not a blocker; (2) confirm the proposed tool is absent from the plugin (canonical .opencode/plugins/uf-workflow/ plus the internal/scaffold/assets mirror) so the issue is not already done; (3) treat pre-migration token-savings estimates as stale metadata to refresh at plan time per DE-004, not as invalidating the issue; (4) confirm the intent-level proposal orchestrates rather than duplicates the granular git_* tools that the intermediate change introduced.
