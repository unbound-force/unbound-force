---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:22Z
identity: instruction-lint-20261107T202322-jay-flowers
tier: draft
---

When a wrapper decouples where a tool lives from what it operates on (unbound-force's .uf/scripts/instruction-lint uses UF_PREFLIGHT_PROJECT_ROOT for the linter + node_modules and --root "$PWD" for the linted tree, so soft-gate merge-base baselines and the primary checkout share toolchain resolution), the tool's machine-parseable output must embed the input scope it was computed against — resolved root, filter scope, tool version. Without it, a captured --json blob is unattributable: baseline-run and primary-run output differ only in file counts, and log-based triage cannot tell which tree produced the findings. Consumers that wrap the invocation (pre-flight tools, agent personas) attach context at their layer, which hides the gap until someone reads raw saved output. Add the root and version to the JSON envelope at emission time — one field each makes captures self-describing per Constitution Principle III (observed on PR #958's instruction-lint --json envelope, .opencode/lib/instruction-lint.ts main()).
