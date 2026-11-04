---
tag: path-safety
author: jay-flowers
category: gotcha
created_at: 2026-11-04T15:01:31Z
identity: path-safety-20261104T150131-jay-flowers
tier: draft
---

Hardening a path-checked read into an atomic one in Node/Bun tooling (review-dispatch readDispatchArtifact and any similar reader): fs.constants.O_NOFOLLOW is POSIX-only — undefined on Windows, where `flags | undefined` coerces to `flags | 0` and silently drops the symlink guard. The .opencode plugin suite tests under vitest/Node while opencode hosts plugins under Bun, so a green test suite does not prove the flag is defined or honored in production; and the repo carried no live O_NOFOLLOW precedent, with the recorded temp-file-security learning simultaneously recommending "O_NOFOLLOW or equivalent" and claiming Node does not expose it directly. Safe shape: feature-detect O_NOFOLLOW, fall back to the existing lstat pre-check where it is unavailable, fstat the opened handle so the size bound and the read share one inode, map ELOOP to the same fail-closed error code the lstat path produced (the symlink test asserted the code, not the message), and use a bounded read (limit+1 bytes) to also close the append-growth race. Mirror the change into internal/scaffold/assets/ (byte-parity drift tests) and land any pending size-bound test against the post-fix reader, since fstat preserves the stat-before-read sparse-truncation test strategy.
