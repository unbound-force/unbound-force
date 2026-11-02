---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:16:08Z
identity: issue-triage-20261101T221608-jay-flowers
tier: draft
---

A documentation-gate tracking issue is triage-verifiable by construction when it records an explicit disposition for every documentation surface the gate names — CHANGELOG.md, AGENTS.md, docs pages, README.md, and content coverage — including justified "no change needed" entries. Verified while triaging unbound-force issue #856 (port-uf-rebase-pr docs impact, 2026-11-01): every disposition cross-checked green against repository state (AGENTS.md PR Review Commands row and the 33-file command-count annotation, docs/usage.md Quick Reference row, docs/architecture.md Review Commands row, CHANGELOG Unreleased/Added entry whose final line carried Spec/Fixes/Docs/website references, scaffold mirror registered in expectedAssetPaths in scaffold_test.go), and the README no-change disposition carried its rationale (no per-command reference table exists; README describes workflows at prose level). Triage heuristic for docs-gate issues: grep each claimed disposition against the repo — an issue whose dispositions are all verifiable is objectively VALID with no author clarification needed, while vague dispositions ("docs updated as needed") or missing no-change rationales signal a gate that was papered over rather than evaluated.
