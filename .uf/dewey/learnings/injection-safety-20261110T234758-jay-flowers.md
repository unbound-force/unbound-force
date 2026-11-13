---
tag: injection-safety
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:47:58Z
identity: injection-safety-20261110T234758-jay-flowers
tier: draft
---

A template-scan regression gate for heredoc payload writes must cover the full bash delimiter quoting grammar and every instruction-prose surface agents execute. PR #1024's TestCommandContracts_RejectShellHeredocPayloads matches only bare and single/double-quoted delimiters (backslash-escaped <<\\EOF and ANSI-C-quoted <<$'EOF' evade it, per learning/shell-injection-20261110T233152-jay-flowers) and scans only .opencode/commands/ plus the scaffold mirror, leaving .opencode/agents/ and .opencode/skills/ unguarded. Verified zero live heredocs across all four surfaces on 2026-11-10, so the gap is latent — extend the regex and the scanned directory list in the tool-extraction follow-up.
