---
tag: git-tool-hardening
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:31:18Z
identity: git-tool-hardening-20261103T213118-jay-flowers
tier: draft
---

Agent-facing git push hardening pattern (validated in unbound-force PR #948, git_push / git_push_force_with_lease rewrite, issue #947): (1) validate ref-name arguments with a schema that rejects leading `-` (option injection), whitespace, and `:` (refspec separator) — argv-array execution prevents shell injection but NOT option injection, and the colon check also blocks every URL form; (2) refuse the push unless `symbolic-ref -q HEAD` equals `refs/heads/<branch>` (BRANCH_MISMATCH, fail-closed on detached HE  [… truncated at capture; final clause unrecoverable — recorded 2026-11-07].
