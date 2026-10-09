---
tag: opsxoptional-profile-models
author: jay-flowers
created_at: 2026-10-04T19:10:22Z
identity: opsxoptional-profile-models-20261004T191022-jay-flowers
tier: draft
---

When implementing OpenSpec changes that involve TypeScript plugin code and Go scaffold assets simultaneously, the scaffold copy at `internal/scaffold/assets/opencode/plugins/review-dispatch/index.ts` must be kept byte-identical to the live plugin at `.opencode/plugins/review-dispatch/index.ts`. Any divergence between the two causes scaffold drift test failures in `internal/scaffold/`. During implementation of the optional profile models feature, the scaffold copy initially had a simplified `catch {}` block for the override file read that silently swallowed all errors, while the live plugin had proper ENOENT checking. This was discovered during code review and fixed. The pattern is: always apply changes to both files simultaneously, not sequentially, and verify with `make check` which runs scaffold tests that validate byte-identical parity.
