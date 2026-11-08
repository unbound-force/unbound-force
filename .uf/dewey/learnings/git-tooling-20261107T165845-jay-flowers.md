---
tag: git-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:58:45Z
identity: git-tooling-20261107T165845-jay-flowers
tier: draft
---

git's valid_remote_nick rejects a remote name through three clauses — empty, is_dot_or_dotdot (`.` and `..`), and directory separators — but the RemoteNameSchema derived from it encoded only the slash clause, so `remote: "."` and `remote: ".."` passed validation and git resolved them as the current and parent repository, letting a lease-less git_push target a neighboring repo's hooks. When a zod schema's doc comment names an upstream validation function as its model, encode that function's full rejection set and add an INVALID_INPUT test row per rejected value. The complete remediation for a tightened tool-argument schema in this repo touches four surfaces: the schema refine, the INVALID_INPUT test rows, the byte-identical scaffold mirror under internal/scaffold/assets/opencode/ (drift tests enforce identity), and every artifact that restates the rule (the pinning openspec delta and the CHANGELOG entry) — otherwise a future re-implementation from those texts reintroduces the gap.
