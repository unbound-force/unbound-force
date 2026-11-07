---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-05T18:34:53Z
identity: security-20261105T183453-jay-flowers
tier: draft
---

A git push tool parameter validated as a "remote nickname" by character class (rejecting `/`, whitespace, `:`, leading `-`) is not yet a configured-remote check: git resolves any unconfigured push target as a URL, and scheme-less, colon-less tokens — `..`, `.`, or a bare name matching a cwd-relative directory — are treated as local repository paths. The `/` rejection added in PR 948 round-2 (learning security-20261105T163851) closes absolute and nested relative paths but not these. Verify nickname-ness semantically before pushing (`git remote get-url <remote>`, failing closed on error) and keep the schema as a fast pre-filter; otherwise `remote: ".."` pushes the branch to the parent-directory repository and runs its receive hooks with the pusher's privileges.
