---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-15T01:40:32Z
identity: spec-authoring-20261115T014032-jay-flowers
tier: draft
as_of: "unbound-force@7f44cba6be6f30f6c3cec406daab342434cc1474"
---

Before drafting Impact/design/tasks for a change that modifies existing plugin tools, grep the tool names to pin the implementation location. The harden-spec-review-gate artifacts guessed "Go code under internal/… and/or .opencode/lib/…" and branched on "if feasible in Go", while consolidate_dispatch/dispatch_review_runs are TypeScript plugin tools in .opencode/plugins/review-dispatch/index.ts (registration ~:9957) and checkpoint/get_progress live in .opencode/plugins/uf-workflow/index.ts. The internal/ Go matches were only scaffold prose-pinning tests, and the research task inherited search paths that omit the plugins directory entirely.
