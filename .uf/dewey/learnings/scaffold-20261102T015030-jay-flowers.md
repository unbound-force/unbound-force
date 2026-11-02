---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-02T01:50:30Z
identity: scaffold-20261102T015030-jay-flowers
tier: draft
---

extractManagedRegion in internal/scaffold/scaffold.go (init-override-stubs-gitignore-sync, design D7): the returned byte range MUST include the terminator line's line break. Reason: managed-block constants end with a trailing newline (gitignoreBlock's raw string has its closing backtick on its own line), and SC-FR-013/SC-FR-014 require the extracted region to equal the block constant byte-for-byte. If end excludes the terminator's newline, a perfectly current file compares as differing, every run reports upgraded, and each in-place rewrite (content[:start] + canonical + content[end:]) re-appends the original newline that was excluded from the replaced region — unbounded newline growth across runs. Rule of thumb: when a managed-region helper backs byte-identical dispositions and in-place rewrites, the region must span exactly the bytes the canonical constant would occupy, including its trailing line break; clamp end to len(content) when the terminator is the final line with no break. Delimiter comparison is line equality with trailing-\r normalization (CRLF tolerance) so a CRLF marker line still matches and never falls back to the append path (duplicate-block prevention); on CRLF files the region keeps \r\n breaks, differs from the LF canonical, and converges to unchanged after one upgrade.
