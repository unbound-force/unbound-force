---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:26:10Z
identity: instruction-quality-20261110T212610-jay-flowers
tier: draft
---

instruction-lint --changed scopes findings to files changed since merge-base(--base, HEAD) of the working tree. In a dispatched PR review whose checkout stays on main, the delta resolves to unrelated main-side files (e.g. uf.address-feedback.md), not the PR's changed instruction files — first-hand change-scoped linter evidence requires the head checked out, which the pre-flight's detached worktrees at head and base provide. Treat the mismatch as Step 1 evidence degradation: report one LOW note and rely on the pre-flight head/base comparison; do not switch branches mid-dispatch (shared-session mutation).
