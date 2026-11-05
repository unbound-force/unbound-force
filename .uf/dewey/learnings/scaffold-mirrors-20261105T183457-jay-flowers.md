---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:34:57Z
identity: scaffold-mirrors-20261105T183457-jay-flowers
tier: draft
---

In manifest-delivered diff reviews, the git index line ("index <old-blob>..<new-blob>") at the top of each file's diff header proves canonical/scaffold-mirror byte parity without retrieving both chunks: identical blob hashes across the mirror pair are cryptographic byte-identity, cheaper than running cmp or reading both 256 KiB chunks. Verified on PR 948 round 3 for all four command mirrors (uf.finale 5888c401..c74b4bfb, uf.fix-ci 08ddb98f..ed5b8428, uf.unleash d4d85ac6..ebce08a3, uf.rebase-pr 5d423725..ab636ab4) and the uf-workflow plugin mirror (ea7dbbe1..3bb790b3); the same technique showed the legacy .opencode/skill/ and new .opencode/skills/ SKILL.md share blob 0f9560f6, i.e. a pure content-preserving move despite asymmetric +78/-76 manifest counts.
