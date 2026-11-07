---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:07:12Z
identity: security-20261107T160712-jay-flowers
tier: draft
---

Developer-CLI tools that echo untrusted file content need control-character sanitization on the plain-text output path even when the JSON path is safe: JSON.stringify escapes ESC/C0 bytes for free, but template-literal interpolation of file-derived values (frontmatter fields, regex-matched path tokens) passes ANSI/OSC sequences straight to the terminal, and text output is often what CI logs and review prompts quote. Sanitize once at the print site — strip or visually escape C0/C1 control sequences except newline and tab — rather than at every emit site, so every future message inherits the boundary. Companion gap in the same class: a lint-type tool whose scope is derived from the working directory should treat a zero-lintable-file run as could-not-run (exit 2), not clean, mirroring the fail-closed rule it applies to filter arguments like --only; running from the wrong directory otherwise reports success with an empty scope.
