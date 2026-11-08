---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:25:03Z
identity: error-handling-20261106T202500-jay-flowers
tier: draft
---

A CLI that documents "exit 1 = findings" is silently violated by uncaught crashes, which node also reports as exit 1 — and in ESM, static imports resolve and evaluate before main() runs, so a missing or corrupted dependency (ERR_MODULE_NOT_FOUND) cannot be caught by a try/catch in main. Exit-code hygiene for lint-type tools needs the whole chain: (a) a distinct could-not-run exit (2, matching the usage-error convention) enforced at the `import.meta.main` entry boundary for runtime exceptions; (b) a dependency surface the manifest actually declares, so a missing module is actionable rather than phantom; (c) a wrapper that pre-checks the entry file, the node_modules directory, and each dependency's package dir before exec'ing node, exiting 2 with the install remedy; and (d) every entry point routed through that wrapper — the contract is only as strong as its weakest entry point, and a Makefile target or agent persona invoking `node` on the library directly reintroduces the crash-as-findings path (pre-flight consumers like uf_preflight_declared_checks classify pass as exitCode === 0 and record the raw code, so exit 2 stays diagnosable without a new status enum). Regression tests must run the real entry points without node_modules: integration tests that symlink node_modules into scratch trees mask the fresh-checkout crash path entirely.
