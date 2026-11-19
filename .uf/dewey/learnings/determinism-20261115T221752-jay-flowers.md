---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-15T22:17:52Z
identity: determinism-20261115T221752-jay-flowers
tier: draft
as_of: "unbound-force@a8f038d3b312a581c6b9d9d1adf637883118d458"
evidence:
  - ".uf/artifacts/dispatch/ed55ae43-93dd-4d32-a7ee-085d3e1c8eb3.json"
  - ".opencode/commands/gaze.md"
  - ".opencode/commands/gaze-fix.md"
  - "openspec/changes/idle-timeout-aborts-children-waiting-on/design.md"
---

DE-001 adjudication for strategy-naming guidance (verified on the idle-timeout-aborts-children-waiting-on council review, branch opsx/idle-timeout-aborts-children-waiting-on): instruction text that names a strategy ("run gaze crap in the background and poll the output file") but prescribes none of its mechanics — no polling interval, output path, completion condition, or output template — and frames the choice as applicability judgment does not violate DE-001. The twice-test fails on the decision step itself (background-vs-synchronous varies with expected runtime and scope), and there is no fixed procedure to extract because the mechanics were never specified. Verify with the owner lookup (node .opencode/lib/tool-inventory.ts --json --match "..."): when no registered tool performs the named strategy (e.g. only uf_finale_ci_poll, which polls PR CI rather than local background processes), a design's "migrate this guidance when a launch-and-poll tool is registered" clause is a conditional forward trigger, not an owed extraction — distinct from the unbound-follow-up violation pattern, which fires only when the added prose IS mechanical (fetch + condition-gate + selection + prescribed output) and defers its own extraction without a tracked issue or tasks.md entry.
