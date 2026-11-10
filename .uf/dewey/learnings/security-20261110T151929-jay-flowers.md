---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:19:29Z
identity: security-20261110T151929-jay-flowers
tier: draft
---

lstat-based symlink refusal on the final file and final directory does not guard a path whose intermediate components are symlinks: feedbackStatePath (unbound-force PR #1007, .opencode/lib/feedback-replies.ts) builds `.uf/feedback/pr-<N>/state.json` lexically with path.resolve(), so a symlink planted at `.uf` or `.uf/feedback` is followed before the lstat checks run — the final-dir lstat sees a real directory inside the target and the write lands outside the intended tree. When the threat model for refusing symlinks includes a planted link (untrusted checkout, shared runner), walk every path component with lstat, or compare the realpath of the resolved parent directory against the expected workspace prefix, instead of checking only the last two components.
