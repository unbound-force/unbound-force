---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-10T16:54:55Z
identity: security-20261110T165455-jay-flowers
tier: draft
---

lstat-based symlink refusal only checks the final path component: lstat of a deep path resolves intermediate symlinks, so a symlink planted at an ancestor directory (.uf or .uf/feedback in PR #1007's feedback cache) redirects both reads and atomic-rename writes past a guard that proudly refuses "symlinked targets". Tests reinforce the blind spot when they symlink only the final file or the immediate directory. For cache-directory controls, resolve the real path once (fs.realpath) and verify it stays under the expected root, or lstat every component from the root down; a final-component check alone provides substantially less protection than its error message claims.
