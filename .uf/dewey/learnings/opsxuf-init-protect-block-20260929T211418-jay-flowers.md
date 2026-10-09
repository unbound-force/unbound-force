---
tag: opsxuf-init-protect-block
author: jay-flowers
category: pattern
created_at: 2026-09-29T21:14:18Z
identity: opsxuf-init-protect-block-20260929T211418-jay-flowers
tier: draft
---

When adding `<protect>` blocks to long slash command files (1000+ lines), place the opening tag immediately after the `## Instructions` heading and the closing tag before the final sections (like `### Next Steps` or `### Post-Write Verification`). This ensures the core workflow steps survive context compaction while keeping metadata sections visible. For idempotency, check for existing `<protect>` tags between `## Instructions` and the first step heading before insertion. When a command modifies other files (like /uf.init does), update insertion logic to detect existing `<protect>` blocks in target files and place new content inside them when present, rather than creating duplicate protection structures.
