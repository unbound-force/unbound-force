---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-02T13:30:07Z
identity: scaffold-20261102T133007-jay-flowers
tier: draft
---

Seeding comment-only YAML stubs from the Go scaffold into per-developer config files consumed by TypeScript loaders is safe only because both loaders treat YAML null as a valid no-op: .opencode/lib/model-matrix.ts parseModelMatrix (null/undefined document resolves to the canonical no-op matrix) and .opencode/plugins/review-dispatch/index.ts parseReviewMatrixOverride (null/undefined override returns early — "empty override is valid"). When a seeder in one language writes config consumed by a loader in another, the no-op safety property lives in the loader, not the seeder — verify the loader's null/empty handling in production code before relying on comment-only seeding. The contract is closed from both sides by pinning the stub content byte-exact in seeder tests (any stub edit forces a test update) plus the loaders' independent null guards.
