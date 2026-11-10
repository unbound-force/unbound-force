---
tag: input-validation
author: jay-flowers
category: pattern
created_at: 2026-11-07T20:41:07Z
identity: input-validation-20261107T204107-jay-flowers
tier: draft
---

When an input guard is syntactic by necessity — a zod schema cannot verify that a remote name is a *configured* remote — the guard's doc comment must record the accepted residual at fix time: what the guard does not close (bare separator-free tokens colliding with cwd-relative git-repo directories resolve as local-path push targets), why it is accepted (every consumer pins the parameter to structured values, tools default to a fixed remote), and what the complete close is (`git remote get-url <name>`, failing closed for unconfigured names, keeping the schema as a fast pre-filter). Across unbound-force PR #948 rounds 2-5, RemoteNameSchema's comment narrated the local-path failure mode for the forms it rejected ('/', '.', '..') but never stated its boundary, so it read as a completeness claim and each review round re-discovered another local-path form behind it. An unwritten residual is re-litigated every review round; a written one is an auditable acceptance. Corollary: the note must land in the byte-identical scaffold mirror under internal/scaffold/assets/opencode/ as well (drift tests enforce identity), and in any spec delta that restates the validation rule, or a future re-implementation from those texts reintroduces the gap.
