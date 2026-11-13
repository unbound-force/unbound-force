---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:45:00Z
identity: security-20261110T224500-jay-flowers
tier: draft
---

When extending TestCommandContracts_RejectShellHeredocPayloads (PR #1024, issue #1011), add .opencode/agents/ and .opencode/skills/ to the scanned directory list — 41 executable bash blocks live there today (including divisor-curator.md, which composes and posts issue content), and a reintroduced heredoc payload write in those files would not fail the gate (verified zero current instances repo-wide on 2026-11-10). Detection nuance: the regex `<<-?\s*['\"]?[A-Za-z_][A-Za-z0-9_]*['\"]?` catches quoted, unquoted, and `<<-` dashed heredocs while correctly skipping herestrings (`<<<`), workflow markers (`>>> MANDATORY GATE <<<`), and conflict-marker prose, but it misses escaped delimiters (`<<\EOF`, `<< \EOF`); add an optional backslash to the prefix class if that exotic form ever needs coverage.
