---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-05T14:45:27Z
identity: scaffold-mirrors-20261105T144527-jay-flowers
tier: draft
---

Scaffold mirror sync (canonical .opencode/* files vs internal/scaffold/assets/opencode/* copies) can be verified from a review diff alone: byte-identical file pairs carry identical git blob hashes in each diff section's "index <old>..<new>" header line. Comparing those index lines across the two sections proves or disproves sync without reading either file — used to verify all five mirrored pairs in PR #948 (commands uf.finale/uf.fix-ci/uf.rebase-pr/uf.unleash plus the uf-workflow plugin).
