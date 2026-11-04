---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:01:34Z
identity: review-dispatch-20261104T150134-jay-flowers
tier: draft
---

Hardening Node file readers against path-swap TOCTOU (PR 949, RD-FR-057 readDispatchArtifact): an lstat(path) guard followed by readFile(path) is two path resolutions — a rename or symlink swap between them bypasses both the regular-non-symlink and the size guards, so the fail-closed contract is enforced against a possibly different inode. Fix pattern: open with O_NOFOLLOW (ELOOP rejects a symlinked final component at open time), fstat the returned FileHandle (isFile + size on the opened inode), and read from the handle (fh.readFile) so guards and read share one inode. Gotcha: fs.constants.O_NOFOLLOW is POSIX-only — undefined on Windows — and `flags | undefined` silently coerces to 0, dropping the guard with no error; branch on flag availability and keep the lstat path where the flag is absent. Preserve errno mappings (ENOENT -> not_found, others -> invalid_artifact): the plugin-suite tests assert error codes via toMatchObject, not messages, so an ELOOP surfacing through the existing catch keeps them green. Any fix to a scaffolded plugin asset must land in both the .opencode/ original and its byte-identical internal/scaffold/assets/ mirror or drift detection fails.
