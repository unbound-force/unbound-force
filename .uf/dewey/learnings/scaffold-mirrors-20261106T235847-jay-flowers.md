---
tag: scaffold-mirrors
author: jay-flowers
category: pattern
created_at: 2026-11-06T23:58:47Z
identity: scaffold-mirrors-20261106T235847-jay-flowers
tier: draft
---

Scaffold-mirror parity can be verified at review time without reading either file: when a changeset touches both a canonical .opencode/ asset and its internal/scaffold/assets/opencode/ mirror, compare the diff's `index <pre>..<post>` header lines — identical post-image blob hashes prove byte-identity cryptographically (validated in unbound-force PR 948: package.json 05467103..b531b24c in both locations, plugin index.ts ea7dbbe1..3bb790b3 in both, package-lock.json 00000000..4c66cfd6 canonical new-file vs ea7f897b..4c66cfd6 mirror). A single grep of the patch file for the post-blob hash surfaces every mirror pair. Combine with the repo's drift-detection tests (make check) as the enforcement evidence; the blob-hash comparison is the reviewer-side confirmation.
