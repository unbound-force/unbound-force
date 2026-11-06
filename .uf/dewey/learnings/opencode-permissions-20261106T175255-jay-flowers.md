---
tag: opencode-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-06T17:52:55Z
identity: opencode-permissions-20261106T175255-jay-flowers
tier: draft
---

OpenCode bash permission allow rules are glob patterns matched over the command string; a trailing `*` (e.g. `"git diff*": "allow"`) matches every argument suffix, including write-capable flags such as `git diff --output=<path>`, which writes a file and bypasses `edit: deny`. Hardening rules for read-only reviewer personas: (1) never use trailing-wildcard allows for commands that have write-capable flags — use exact-match allow entries so unmatched invocations fall through to the `"*": "deny"` default, which is independent of OpenCode's deny/allow precedence; (2) a deny rule like `*--output*` has two gaps — deny-over-allow precedence must be verified (a `"*": "deny"` entry coexisting with allows proves resolution is not simple deny-wins), and picomatch-style `*` does not cross `/`, so slash-bearing paths need `**--output**`; (3) audit every allow entry's command for write-capable flags, including script CLIs allowed via `node <script>*` (instruction-lint.ts and tool-inventory.ts are write-free; review-diff.ts is not). Recurred across stacked PRs #958 (divisor-linguist.md:11) and #959 (divisor-determinism.md:12).
