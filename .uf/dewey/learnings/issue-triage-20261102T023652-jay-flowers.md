---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:36:52Z
identity: issue-triage-20261102T023652-jay-flowers
tier: draft
---

Documentation-tracking issues in the unbound-force org accumulate scope updates from related OpenSpec changes via comments — read the full comment thread before assessing duplication or staleness. Issue unbound-force/unbound-force#710 (uf-finale plugin docs) demonstrates the pattern: the later fix-tui-git-output-leak change threaded its documentation requirements (envelope-only tool output, uf_finale_stage {fileCount, files} return shape, silent uf_finale_ci_poll) into the existing open docs issue as a scope-update comment, with the cross-repo website tracking issue (unbound-force/website#306) referenced inline. Sibling docs issues from the same deterministic-tool migration program (#739 unleash migration + unleash summary, #810 base_state gate) document distinct tool/artifact scopes and are siblings rather than duplicates — distinguish them by the specific tools each issue enumerates (e.g., #710's 8-tool list excluding uf_finale_base_state), not by shared "uf.finale docs" keywords.
