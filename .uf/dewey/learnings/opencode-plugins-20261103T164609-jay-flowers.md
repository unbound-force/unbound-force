---
tag: opencode-plugins
author: jay-flowers
category: gotcha
created_at: 2026-11-03T16:46:09Z
identity: opencode-plugins-20261103T164609-jay-flowers
tier: draft
---

The OpenCode plugin host does not enforce a tool's declared zod `args` shape before `execute()` runs — the declaration is documentation-grade for object-valued args. An arg that execute() destructures (e.g., the uf-workflow checkpoint tool's `phase`) must be re-validated inside execute through the plugin's shared `parseToolInput` seam; otherwise malformed input such as a bare `phase: 5` surfaces as downstream type garbage (`requires step "implement-phase-undefined"`) instead of a shape-naming `INVALID_INPUT` error.
