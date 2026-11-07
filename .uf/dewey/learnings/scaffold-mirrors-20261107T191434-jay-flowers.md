---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:14:34Z
identity: scaffold-mirrors-20261107T191434-jay-flowers
tier: draft
---

Manifest-mode review diffs that carry both a canonical file and its embedded scaffold mirror (e.g. .opencode/plugins/uf-workflow/index.ts and internal/scaffold/assets/opencode/plugins/uf-workflow/index.ts, or the command/package mirrors) prove byte-identity for free: the `index <old-blob>..<new-blob>` hash pairs in the two diff headers are content-addressed, so identical pairs mean identical bytes. A reviewer verifies mirror parity without retrieving the mirror's chunk at all — retrieve one copy, compare the blob-hash pairs in the two diff headers. Applied on PR 948: canonical and mirror showed identical pairs (uf.rebase-pr.md 5d423725..ab636ab4, uf-workflow/index.ts ea7dbbe1..60661ef4, package.json 05467103..b531b24c, package-lock.json ..4c66cfd6), and the in-repo command-contract and drift tests independently enforce the same parity.
