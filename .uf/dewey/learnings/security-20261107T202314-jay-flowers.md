---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:14Z
identity: security-20261107T202314-jay-flowers
tier: draft
---

When adding control-character/terminal-injection escaping to a CLI, enumerate every output site, including stderr diagnostics — not just the findings/report stdout paths. V8's JSON.parse SyntaxError messages embed a snippet of the parsed source and the offending token, so an error-report line like `cannot read config ${path}: ${e.message}` re-introduces raw file content on stderr even after every stdout site is escaped. Route error messages through the same escape helper at all io.err sites, and pin the coverage with a test that plants a control character in the failing input and asserts the escaped form on both streams. Precedent: PR #958's instruction-lint escaped its stdout print sites (escapeControlChars + unit and integration tests) but left the config-parse stderr path raw.
