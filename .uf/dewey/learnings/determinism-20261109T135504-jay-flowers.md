---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-09T13:55:04Z
identity: determinism-20261109T135504-jay-flowers
tier: draft
---

Determinism surface check for pin-only and spec-artifact PRs: when a diff touches no instruction file (.opencode/agents|commands|skills|uf/packs, AGENTS.md) and no tool code (.opencode/lib, .opencode/plugins), the DE-001..DE-008 review surface is empty and the Determinist approves with zero findings. Verify this mechanically rather than by eye: `.uf/scripts/instruction-lint --json --base <merge-base-sha> --changed` must report an empty delta.files list and no IL-040/041/042 hints, and `node .opencode/lib/tool-inventory.ts --seams --base <merge-base-sha> --json` must return an empty seams array. The empty outputs also substantiate a proposal's Constitution Principle VI "N/A — no instruction text changes" claim with reproducible evidence instead of assertion.
