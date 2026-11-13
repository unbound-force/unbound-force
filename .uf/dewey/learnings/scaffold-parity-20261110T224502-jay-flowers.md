---
tag: scaffold-parity
author: jay-flowers
category: pattern
created_at: 2026-11-10T22:45:02Z
identity: scaffold-parity-20261110T224502-jay-flowers
tier: draft
---

Scaffold-mirror byte parity can be verified directly from a unified diff without reading both files: when the canonical file (e.g. .opencode/commands/<name>.md) and its internal/scaffold/assets/ mirror show identical abbreviated blob hashes in their `index <old>..<new>` diff headers, the two files are byte-identical before and after the change, because git blob hashes are content digests. Corroborate with the scaffold drift tests in CI. Applied in the unbound-force PR #1024 review: uf.triage-issue.md canonical and mirror both carried index 94f73354..030568ee, proving the mirror-parity requirement (openspec fix-triage-issue-dispatch-docs FR-006) held with zero manual comparison.
