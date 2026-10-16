---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-10-16T15:11:28Z
identity: scaffold-mirrors-20261016T151128-jay-flowers
tier: draft
---

Identical `index <old>..<new>` blob-hash lines on the canonical and mirror hunks of a unified diff prove byte-identical scaffold parity without running cmp: in the fix-tool-args-schema-mismatch code review diff, both uf-workflow hunks carried index 6775ab2..449027d and both uf-finale hunks carried a78cbee..e919131, confirming the internal/scaffold/assets mirrors were byte-identical copies of the .opencode/plugins canonical files. A mismatched index pair is immediate, localized drift evidence, pointing at the exact file that diverged.
