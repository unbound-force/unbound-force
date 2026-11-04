---
tag: untrusted-output-rendering
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:00:12Z
identity: untrusted-output-rendering-20261103T210012-jay-flowers
tier: draft
---

When rendering untrusted output into a CommonMark fenced block, sizing the fence past the content's longest backtick run is necessary but not sufficient: the closing fence must start on its own line, so the renderer must normalize a trailing newline onto the body (output.endsWith("\n") ? output : output + "\n") before emitting the closer. Without it, output lacking a trailing newline glues the closing fence onto the last content line ("abc```"), which is not a valid closing-fence line — the block never clos
