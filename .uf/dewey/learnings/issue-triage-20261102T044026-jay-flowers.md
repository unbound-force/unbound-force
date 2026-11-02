---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:40:26Z
identity: issue-triage-20261102T044026-jay-flowers
tier: draft
---

Documentation-gap triage for a command whose name is a prefix of another command requires content-level verification, not keyword matching. Verified on unbound-force/unbound-force issue #583 (`uf gate` compliance command): all 13 "uf gate" grep matches in docs/ are actually `uf gateway` — the LLM reverse proxy, a different command with its own cli-reference.md section (lines 351-408) — and the issue tracker's nine fallback duplicate candidates (#819, #862, #810, #856, #825, #908, #747, #838, #740) are gate-mechanism or other-command docs issues, none about the CLI command. When verifying a claimed documentation gap or ruling out duplicates for such commands, grep the command token together with its distinctive syntax (flags, subcommands, exit-code tables) and confirm each match documents the target command's CLI surface rather than a lexical neighbor.
