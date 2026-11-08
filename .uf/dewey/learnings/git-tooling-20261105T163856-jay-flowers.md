---
tag: git-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:38:56Z
identity: git-tooling-20261105T163856-jay-flowers
tier: draft
---

Reusing one zod schema for both `branch` and `remote` fields of a git push tool under-validates the remote: git reinterprets a slash-bearing push target that is not a configured remote as a local repository path, so a `remote` value like `/tmp/repo` passes branch-shaped validation and silently copies the branch's commits to a local path. Split validation by git identifier semantics — branch names legitimately contain `/` (feat/x, quick/<name>), remote names never need it — and reject `/` only in a dedicated remote schema. A blanket `/` rejection in the shared PushRefNameSchema would break every slash-named branch push, including the test suite's own feat/x fixtures.
