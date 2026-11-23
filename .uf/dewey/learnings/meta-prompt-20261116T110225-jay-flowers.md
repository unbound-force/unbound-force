---
tag: meta-prompt
author: jay-flowers
category: pattern
created_at: 2026-11-16T11:02:25Z
identity: meta-prompt-20261116T110225-jay-flowers
tier: draft
as_of: "unbound-force@474f5aee31d7e0bc4583635cccfeec6030277b33"
---

When instruction prose branches on tool-envelope fields (e.g. `timeout.resume.files_modified`, `timeout.child_active`, `remaining_ms`, `phase_checkpoint_line`), no instruction-lint rule verifies those field names against the tool layer's schema. During instruction review, grepping the plugin/lib source for each cited field name confirms the branch conditions are actually checkable by the executing agent — field-name drift between command prose and tool schema silently makes the instruction unexecutable. In PR #1157's uf.unleash.md per-task budget contract, all cited envelope fields verified present in agent-executor.ts / uf-workflow/index.ts, and the plugin's phase_checkpoint_line rendering (budget clause exactly when remaining_ms was recorded) matched the prose claim verbatim.
