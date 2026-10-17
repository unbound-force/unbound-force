---
tag: symlink-defense
author: jay-flowers
category: reference
created_at: 2026-10-16T11:50:34Z
identity: symlink-defense-20261016T115034-jay-flowers
tier: draft
---

The persist_unleash_summary tool extends the checkpoint writeState symlink defense pattern by adding lstatSync checks on both the parent directory AND the target file before writing. The checkpoint writeState only guards the parent directory. While the TOCTOU gap between lstat and writeFileSync is negligible in a single-user local context, the dual-check pattern is a good defense-in-depth practice for any filesystem write tool in the uf-workflow plugin. Future tools that write to user-influenced paths should follow the persist_unleash_summary pattern (directory + file symlink check) rather than the older writeState pattern (directory only).
