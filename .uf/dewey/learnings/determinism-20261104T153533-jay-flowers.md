---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-04T15:35:33Z
identity: determinism-20261104T153533-jay-flowers
tier: draft
---

DE-001/DE-005 extraction assessment when a prescribed loop crosses a tool-server boundary: if the loop must call an MCP-server tool (e.g. dewey_store_learning on the Dewey server) that the plugin layer cannot invoke, extraction is infeasible in-repo and the residual is properly classified LOW with a documented retirement path (a future batch-store tool on the owning server), not escalated. Three tests decide the address-feedback disposition for such an item: (1) the AGENTS.md determinism gate binds mechanical procedural prose only 'where an executable tool is feasible' — the feasibility qualifier governs; (2) per learning determinism-20261101T225413, an extraction follow-up is actionable only when the target executable surface exists in the repository — a cross-repo target makes the item AUTHOR-DECIDES, since there is no in-scope branch change to apply and ACCEPT is inapplicable; (3) when the loop implements a spec MUST (PR 949's RD-FR-056 artifact-sourced consumer contract), removing or vagueifying the prescription would contradict the spec — retirement requires a spec amendment alongside the new tool. The correct in-scope action is documenting the accepted residual and retirement path in the feedback reply; keep the exact prescribed prose in the meantime, because deterministic instruction beats vague intent for drift control. Shipped example: PR 949's Lesson storage rewrite, which tool-extracted the read side (read_dispatch_artifact paging) and left only the MCP-boundary store step as prose.
