---
tag: cli-tooling
author: jay-flowers
category: gotcha
created_at: 2026-11-08T21:33:05Z
identity: cli-tooling-20261108T213305-jay-flowers
tier: draft
---

A wrapper that "pins" a flag by prepending it before "$@" (exec node script --root "$PWD" "$@") pins nothing when the wrapped CLI's argument parser lets later occurrences overwrite earlier ones. The pin holds only against omission, not against override: a caller-supplied `--root /other/tree` (or `--config /arbitrary.json`, which can disable the very rules the evidence exists to report) silently wins. Enforce pinning structurally — reject the flag's presence in "$@" with a loop before exec, or make the wrapped CLI error on duplicate flags — and never describe an advisory prepend as an enforced invariant in persona frontmatter comments, because reviewers and guard tests will cite the comment as the security property.
