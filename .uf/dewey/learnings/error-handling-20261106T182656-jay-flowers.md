---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:26:56Z
identity: error-handling-20261106T182656-jay-flowers
tier: draft
---

Node exits 1 on any uncaught exception, so a CLI wrapper whose documented contract says "exit 1 = findings reported" conflates crashes with findings. Reserve exit 2 for could-not-run failures and enforce it at every boundary that can crash before the tool's main() runs. Static ESM imports resolve before main() executes, so module-resolution failures (e.g. an undeclared transitive dependency like `yaml`, resolved via node_modules from @opencode-ai/plugin) cannot be caught by a try/catch in main — guard them in the wrapper (pre-check node_modules before exec, exit 2 with an actionable message) or convert to a dynamic import; runtime exceptions get a top-level try/catch in the `import.meta.main` entry block. Pre-flight consumers like uf_preflight_declared_checks classify pass as `exitCode === 0` and record the raw exit code, so exit 2 keeps crashes diagnosable (rendered as `exit 2` with the message) without requiring a new status enum.
