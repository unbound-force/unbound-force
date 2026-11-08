---
tag: cli-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-08T21:33:04Z
identity: cli-tooling-20261108T213304-jay-flowers
tier: draft
---

A CLI that enumerates a directory and readFileSync-es every entry must apply the same entry-type guard its file reader uses. listDir-style readdirSync returns directories alongside files, so a scripts directory that later gains a subdirectory (or an unreadable entry) throws EISDIR/EACCES on the first such entry; the throw escapes to the shared runMain wrapper and exits the entire CLI with code 2, converting a one-entry layout change into total loss of the tool's evidence for every consumer. Guard at the enumeration site (statSync(abs).isFile() filter, mirroring the fsReader helper in the same module, or a try/catch that skips the entry) rather than relying on runMain's crash conversion — the exit-code contract is for genuine could-not-run conditions, not for predictable filesystem layout variance.
