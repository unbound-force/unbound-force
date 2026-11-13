---
tag: documentation
author: jay-flowers
category: pattern
created_at: 2026-11-10T19:49:49Z
identity: documentation-20261110T194949-jay-flowers
tier: draft
---

When an enumeration in documentation gains an entry, the clauses adjacent to it must gain their matching content in the same edit: docs/osv-transitive-scan.md's Provenance paragraph gained effect (advisory, removal condition) while the neighboring rollback sentence still names only the PR #957 zod/toml pin commit, leaving effect's rollback discoverable only in the CHANGELOG. Adjacent rollback, removal, and verification clauses drift silently when only the list is extended; reviewers should read one sentence past every extended enumeration.
