---
tag: opencode-plugins
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:13:50Z
identity: opencode-plugins-20261102T021350-jay-flowers
tier: draft
---

The fix-tool-args-schema-mismatch migration is complete: all 11 tool factories in .opencode/plugins/uf-workflow/index.ts (8) and .opencode/plugins/uf-finale/index.ts (3) now declare parameters via zod `args`, closing the "legacy tools need migration" state recorded in learning opencode-plugin-sdk-20261015T234420-jay-flowers. Runtime symptom signatures that indicate this failure class has regressed (all silent at the registration site, since @opencode-ai/plugin emits no error for a plain `parameters:` key): the model falls back to bash for git_push; uf_finale_ci_poll returns INVALID_PR_NUMBER despite explicit args; uf_finale_create_pr requires a bash fallback (as PR #733 did); and commit model-name attribution (uf_finale_commit_context / normalize_model_name) goes missing. Detection belongs in registration-surface tests verifying input-schema parity (the DE-007 seam) because the tools remain callable while presenting empty schemas to the model. Documentation trail: unbound-force/unbound-force#732 (root cause), #740 (docs gate record), unbound-force/website#310 (cross-repo sync).
