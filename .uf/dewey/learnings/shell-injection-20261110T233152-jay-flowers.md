---
tag: shell-injection
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:31:52Z
identity: shell-injection-20261110T233152-jay-flowers
tier: draft
---

A quoted shell heredoc suppresses variable expansion but not delimiter breakout: a payload line exactly equal to the delimiter terminates the heredoc early and the remaining payload executes as shell commands, so LLM-composed or issue-derived text must never flow through one. Write composed payloads with a file API or an intent-level tool (mktemp, 0600, programmatic JSON serialization, gh api --input, finally-block cleanup) and pass only the file path through the shell. When guarding the regression with a template scan, cover the full bash delimiter quoting grammar: backslash-escaped (<<\EOF) and ANSI-C-quoted (<<$'EOF') delimiters evade a pattern that matches only bare and single/double-quoted identifiers.
