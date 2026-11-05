---
tag: review-dispatch
author: jay-flowers
category: gotcha
created_at: 2026-11-03T11:10:59Z
identity: review-dispatch-20261103T111059-jay-flowers
tier: draft
---

When adding an input field to an OpenCode plugin tool in .opencode/plugins/review-dispatch/index.ts, the field must land in THREE places or it is not reachable: (1) the internal Zod input schema (e.g. ConsolidateDispatchInputSchema), (2) the tool registration's `args: {...}` shape in createXxxTool (this is what the SDK advertises to calling LLMs and what boundary validation sees), and (3) the tool `description` text. The vitest suite invokes tool execute() directly with hand-built args, so tests written against builders or execute() will pass while the host path silently drops the parameter (PR #929: `detail` on consolidate_dispatch was added to the Zod schema but never to args or the description, making RD-FR-049 summary mode unreachable through the real host).
