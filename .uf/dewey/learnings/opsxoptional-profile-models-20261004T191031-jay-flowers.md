---
tag: opsxoptional-profile-models
author: jay-flowers
created_at: 2026-10-04T19:10:31Z
identity: opsxoptional-profile-models-20261004T191031-jay-flowers
tier: draft
---

For OpenSpec changes that modify plugin code, the test suite in `.opencode/test/` uses `vi.mock` to simulate filesystem calls. When adding override file loading to `loadPolicies`, the mock `readText` implementation in the test `dependencies()` helper must throw ENOENT-coded errors for the override path so the function can distinguish "file not found" from other read failures. Without properly simulating ENOENT, the test sees all read failures as generic errors, causing the override absent test cases to incorrectly trigger INCONCLUSIVE paths.
