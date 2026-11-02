---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:16:15Z
identity: issue-triage-20261101T221615-jay-flowers
tier: draft
---

unbound-force documentation-gate issues (title prefix "docs:", "docs" label) follow a fixed disposition pattern per the AGENTS.md documentation gate: CHANGELOG.md entry under Unreleased/Added with Spec/Fixes/Docs issue references, AGENTS.md structural updates (PR Review Commands table row, Project Structure file-count bump), docs/usage.md Quick Reference row, docs/architecture.md table row, README.md assessment with an explicit no-change rationale when applicable, content-coverage decisions (tutorial yes / blog no, with significance-threshold rationale), and an unbound-force/website tracking reference for the constitution's Cross-Repo Documentation rule. These issues can be triaged objectively: the claimed AGENTS.md table rows and structure counts are directly checkable against the repository's actual AGENTS.md content (e.g., the /uf.rebase-pr row and the 33-files commands count).
