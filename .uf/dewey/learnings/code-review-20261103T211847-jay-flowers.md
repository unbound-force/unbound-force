---
tag: code-review
author: jay-flowers
category: pattern
created_at: 2026-11-03T21:18:47Z
identity: code-review-20261103T211847-jay-flowers
tier: draft
---

Git diff headers carry blob content hashes in the `index <old>..<new>` line. When a changeset ships scaffold or registry mirrors of canonical files (unbound-force ships every .opencode/ file twice under internal/scaffold/assets/ and schemas/ three ways), identical index hashes between the canonical diff and the mirror diff prove byte-identity of both the pre- and post-images without re-reading the mirror content — a cheap, mechanical verification for mirror-heavy changesets delivered in manifest mode, compl
