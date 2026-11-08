---
tag: agent-permissions
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:57:30Z
identity: agent-permissions-20261108T185730-jay-flowers
tier: draft
---

A persona bash-permission hardening convention enforced by a test that reads a single precedent persona regresses silently the moment a second persona with a bash map ships. In PR 959, divisor-linguist.md's bash map gained terminal metacharacter deny globs (last-match-wins prefix globs: *>*, *<*, *|*, *;*, *&, *$(, backtick) with a vitest case reading only divisor-linguist.md — while the same PR added divisor-determinism.md with two trailing-star `node …*` allows, no terminal denies, and a `--match` argument composed from untrusted diff text: exactly the injection surface the denies exist to close. Pattern: whenever a persona gains or edits a permission.bash map, mirror the terminal deny block after the last allow/ask rule, and make the hardening test glob every .opencode/agents/*.md declaring a bash map instead of pinning one file (divisor-curator.md fails the same convention pre-existing).
