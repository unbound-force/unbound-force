---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-16T19:09:56Z
identity: instruction-quality-20261116T190956-jay-flowers
tier: draft
as_of: "unbound-force@067fcc2c0cdadebd00928e76c1fe237973de59e8"
---

Scaffold-mirror byte-identity for an instruction-file PR can be verified with zero working-tree access: when the caller-supplied diff shows the same `index <old>..<new>` blob hashes on both the canonical file (.opencode/...) and its internal/scaffold/assets mirror, both files share one git blob at head — blob hashes are content digests, so identity is proven cryptographically (cross-check with the pre-flight worktree `diff -q` when present). Related operational fact: the instruction linter's `--only` scope rejects scaffold mirrors as non-lintable (exit 2), so lint the canonical file only; mirror parity is owned by drift tests plus the hash proof. Also, when the review checkout sits at the base SHA, `--base BASE --changed` computes an empty delta (merge-base(BASE, HEAD) == HEAD) — scope with `--only <canonical changed files>` for baseline evidence and use the pre-flight's head-SHA run for the new-text evidence. Verified while reviewing PR #1169, where both uf.triage-issue.md copies carried `index 2d73684e..e271fe12`.
