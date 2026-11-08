---
tag: opencode-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:16:52Z
identity: opencode-permissions-20261106T201652-jay-flowers
tier: draft
---

In a stacked-PR chain, a security remediation captured as a retrospective learning in the upper PR does not fix the lower PR's files: the upper branch's cumulative diff against the merge base still carries the unfixed pattern, so each branch must be verified against its head state, not against its learnings. Observed on unbound-force PR #959: the branch's own learnings (opencode-permissions-20261106T175255, agent-permissions-20261106T175259) document that trailing-star bash allows like "git diff*" admit git diff --output (file write bypassing edit:deny) and git difftool --extcmd (arbitrary command execution), and divisor-determinism.md was fixed to exact-match entries — but divisor-linguist.md:11, carried from stacked PR #958, still shipped the trailing-star allow. Pre-PR check: for every security learning that names a file:line, grep the branch's cumulative head state for the offending pattern; a learning without a matching fix in the same tree is an open finding, not a closed one.
