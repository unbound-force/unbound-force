---
tag: cross-command-consistency
author: jay-flowers
category: pattern
created_at: 2026-06-23T19:56:32Z
identity: cross-command-consistency-20260623T195632-jay-flowers
tier: draft
---

When standardizing content across multiple slash command files (e.g., model resolution instructions in finale.md, address-feedback.md, and review-pr.md), the Architect reviewer will flag DRY violations if the instruction text differs in substance — not just formatting — across the files. The fix is to ensure semantically identical instructions use the same numbered step format and include all SHOULD-level behaviors (like user warnings) consistently. Line wrapping may vary due to indentation context, but the content and behavioral contracts must match. Copy-pasting from the most complete version (typically the primary command being enhanced) to the secondary commands ensures consistency.
