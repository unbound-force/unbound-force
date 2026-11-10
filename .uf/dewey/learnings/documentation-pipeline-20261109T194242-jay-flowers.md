---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-09T19:42:42Z
identity: documentation-pipeline-20261109T194242-jay-flowers
tier: draft
---

When a documentation change restates a tool's decision conditions or contract (light-check gates, exit codes, flag semantics), verify the prose against the implementation's exported constants and predicates rather than the tool's own natural-language description — the description can be lossier than the code. On unbound-force PR #1000 (issue #996), the corrected resolve_review_mode light-check docs matched .opencode/plugins/review-dispatch/index.ts exactly — 200-line cap, STRUCTURAL_FILE_STATUSES ("added", "removed", "renamed", "copied"), SENSITIVE_PATH_PATTERNS (.github/, AGENTS.md, constitution.md, .opencode/uf/packs/, auth|token|secret|credential), fail-closed on unreadable inputs — including the "copied" file status that the tool's own prose description omits. A doc correction derived from the description instead of the source would have perpetuated the omission it was fixing. The same check pattern applies to CLI docs: confirm documented flags against the CLI's USAGE string and exit codes against the wrapper's header comment.
