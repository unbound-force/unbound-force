---
tag: git-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:38:54Z
identity: git-tooling-20261105T163854-jay-flowers
tier: draft
---

Git-subprocess tool schemas: `branch` and `remote` push arguments have disjoint grammars. Branch names REQUIRE `/` support (hierarchical refs like feat/x, quick/name), while remote names FORBID it: git's valid_remote_nick rejects directory separators, so a slash-bearing push target can only resolve as a local repository path (URLs are separately blocked by rejecting `:`). Sharing one ref-name guard across both fields leaves the remote field accepting local-path push targets. The pattern-conformant fix: derive a RemoteNameSchema from the shared guard via an additional refine (!value.includes('/')), apply it to every `remote` field, and add the rejected shape (remote: "/tmp/repo") to the INVALID_INPUT table test. Prefer schema rejection over live `git remote` validation: it keeps the failure at parse time (INVALID_INPUT) and unknown-but-well-formed remote names already fail at push time.
