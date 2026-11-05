---
tag: path-safety
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:48:35Z
identity: path-safety-20261105T174835-jay-flowers
tier: draft
---

Feature-detecting a platform-conditional filesystem security flag with `flag ?? 0` silently disables the guarantee the surrounding docstring documents. In review-dispatch's readDispatchArtifact (change review-dispatch-consumer-bounds), the docstring claimed a symlink cannot redirect the read, but constants.O_NOFOLLOW is undefined on Windows, where `?? 0` drops the guard and stat() follows the link; the change's own path-safety learning prescribed an lstat fallback that never shipped. The POSIX-only symlink test cannot catch this — a green suite proves the guard only on the CI platform. Pattern: pair every `flag ?? 0` detection with an equivalent fallback (lstat pre-check rejecting non-regular files, or realpath containment against the canonical store directory, as the sibling-evidence reader already does), and scope docstring security claims to what holds on every supported platform.
