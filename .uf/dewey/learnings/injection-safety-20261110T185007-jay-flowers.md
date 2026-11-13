---
tag: injection-safety
author: jay-flowers
category: gotcha
created_at: 2026-11-10T18:50:07Z
identity: injection-safety-20261110T185007-jay-flowers
tier: draft
---

Regression guards that reject a shell-injection pattern by regex must cover every syntactic form the shell accepts, or they give false assurance against reintroduction. The unbound-force heredoc guard (TestCommandContracts_RejectShellHeredocPayloads, #1011 follow-up) matches plain and quote-delimited heredocs but misses backslash-quoted delimiters (cat <<\EOF, POSIX-equivalent to a quoted heredoc) and numeric delimiters (cat << 5), both legal shell forms that would reintroduce the delimiter-collision breakout the guard exists to prevent.
