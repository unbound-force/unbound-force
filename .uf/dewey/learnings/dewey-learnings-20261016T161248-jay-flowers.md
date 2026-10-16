---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-10-16T16:12:48Z
identity: dewey-learnings-20261016T161248-jay-flowers
tier: draft
---

Dewey learning deduplication is identity-scoped ({tag}-{YYYYMMDDTHHMMSS}-{author}), so two learnings committed under different topic tags with near-identical content both persist and fragment future semantic search — one concept returns two overlapping draft-tier results, and neither supersedes the other. In the fix-tool-args-schema-mismatch change, scaffold-mirrors-20261016T151128 and scaffold-parity-20261016T151133 documented the same diff-index-pair parity verification under different tags. Before committing a batch of learnings produced during a review session, scan the same-session learnings for overlapping content regardless of tag, and either consolidate into one learning or explicitly differentiate each one's scope so a later compile can merge them cleanly.
