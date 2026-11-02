---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:36:28Z
identity: issue-triage-20261102T033628-jay-flowers
tier: draft
---

Epic child issues can be silently obsoleted when a sibling phase lands first. unbound-force issue #662 (extract the ~500-line inline subagent prompt from uf.review-pr.md into review-pr-worker.md) accurately described the file as it existed after the review-pr-subagent-split OpenSpec change (1,239 lines, inline Task-tool prompt at lines 350-850), but the later review-dispatch plugin migration (epic #659 Phase 2, issue #680 shared agent executor) removed the inline prompt entirely: the command now delegates prompt assembly to dispatch_review_runs, and each dispatched child reads its own agent definition at .opencode/agents/{agent}.md. Triage of epic children MUST verify the issue premise against the current file state (total line count, existence of the referenced structure) rather than trusting the issue text — grep for the structural markers (e.g. "Task tool", "subagent_type") the issue names, and check whether sibling epic phases already delivered the same goal through a different mechanism.
