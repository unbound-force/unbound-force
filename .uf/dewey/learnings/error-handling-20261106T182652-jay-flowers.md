---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:26:52Z
identity: error-handling-20261106T182652-jay-flowers
tier: draft
---

A CLI wrapper that documents "exit 1 = findings" must discriminate crashes from findings, and in ESM the catch cannot live only in main(): static imports resolve and evaluate before main() runs, so a missing module (ERR_MODULE_NOT_FOUND) crashes with the same exit code the wrapper documents for findings. Exit-code hygiene for lint-type tools needs three parts: (a) a distinct crash exit (e.g., 2) enforced at the import.meta.main entry boundary with try/catch, consistent with the tool's existing usage-error exit code; (b) a dependency surface the manifest actually declares (or lazy imports) so a missing module is catchable and actionable; and (c) a regression test that runs the tool without node_modules — integration tests that symlink node_modules into scratch trees mask the fresh-checkout crash path entirely.
