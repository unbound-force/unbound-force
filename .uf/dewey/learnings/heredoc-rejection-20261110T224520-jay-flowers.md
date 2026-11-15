---
tag: heredoc-rejection
author: jay-flowers
category: gotcha
created_at: 2026-11-10T22:45:20Z
identity: heredoc-rejection-20261110T224520-jay-flowers
tier: draft
---

The bash escaped-delimiter heredoc `<<\EOF` (and `<<\ 'EOF'`) behaves exactly like a quoted heredoc — no expansion, and the body still terminates at any line equal to the delimiter — so it carries the same delimiter-breakout injection risk that motivated removing quoted heredocs from command templates. Heredoc-detector regexes that require a quote or identifier immediately after `<<` (e.g. `<<-?\s*['\"]?[A-Za-z_][A-Za-z0-9_]*['\"]?`) miss the backslash spelling; allow an optional backslash (`<<-?\s*\\?['\"]?...`). Pin both the recognized spellings (quoted, double-quoted, unquoted, dashed, escaped, here-string via the second `<`) and the non-matches (`>>> MANDATORY GATE <<<` markers, `<<<<<<<` conflict-marker prose) in a table-driven subtest so the covered set is itself regression-tested.
