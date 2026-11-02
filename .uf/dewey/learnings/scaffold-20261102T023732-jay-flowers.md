---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:37:32Z
identity: scaffold-20261102T023732-jay-flowers
tier: draft
---

Closing sentinels for managed blocks in gitignore-syntax files (.gitignore, .git/info/exclude) MUST be comment lines (leading '#'). Two independent reasons from init-override-stubs-gitignore-sync task 1.3 (stealthExcludeEndMarker, design D8 / spec review advisory A2): (1) a non-comment line in such files is a live pattern — git would match paths against the sentinel text itself; (2) line-based parsers like stealthExcludedPaths count every non-comment line, so a non-comment sentinel corrupts the extra-path accounting. General rule for managed-region writers built on extractManagedRegion(content, marker, terminator): the terminator constant doubles as both a parser delimiter and file content, so it must be chosen to be inert in the host file format (comment in gitignore syntax, tag in markup, etc.). Extras appended by companion writers (appendNewUntrackedToExclude) go at end of file after the sentinel, outside the region, so in-place upgrades splice content[:start]+canonical+content[end:] without touching them; rollback that restores pre-write bytes wholesale is structurally unaffected by region shape.
