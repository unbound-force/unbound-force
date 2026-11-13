---
tag: security
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:49:45Z
identity: security-20261110T194945-jay-flowers
tier: draft
---

Heredoc-rejection contract tests should scan templates per line with a regex that requires an identifier after `<<` (e.g. `<<-?\s*['\"]?[A-Za-z_][A-Za-z0-9_]*['\"]?`): this catches quoted, unquoted, and `<<-` dashed heredocs (and here-strings via the second `<`), while NOT false-positiving on `>>> MANDATORY GATE <<<` workflow markers or conflict-marker prose (`<<<<<<<`, `=======`) because those contain no identifier immediately after `<<`. Scope the scan to every directory carrying executable instruction templates — commands alone leaves agents/ and skills/ able to reintroduce the heredoc-payload injection pattern without failing the gate. Verified in unbound-force TestCommandContracts_RejectShellHeredocPayloads (PR #1024, issue #1011).
