---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:55:40Z
identity: scaffold-mirrors-20261107T195540-jay-flowers
tier: draft
---

When reviewing manifest-delivered diffs in a repository with scaffold mirrors (canonical .opencode/ file plus internal/scaffold/assets/ copy), the diff `index <old-blob>..<new-blob>` header lines verify canonical-to-mirror byte-identity without running drift tests: identical after-blob hashes prove the mirror carried exactly the same bytes as the canonical file, and identical before-hashes prove both copies started from the same base. This converts a would-be "trust the drift tests" assumption into direct review evidence and catches accidental mirror divergence even when the reviewer cannot execute the suite. Observed on PR 948: the uf-workflow plugin, all four command files, and both package-lock.json copies each showed matching blob-hash pairs across canonical and mirror (e.g. both lock diffs ending at ..4c66cfd6a), so byte-identity was established mechanically before consulting the test suite.
