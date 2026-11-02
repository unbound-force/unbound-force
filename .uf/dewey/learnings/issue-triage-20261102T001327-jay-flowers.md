---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:13:27Z
identity: issue-triage-20261102T001327-jay-flowers
tier: draft
---

Documentation-companion issues in unbound-force repos that track a PR's documentation obligations can be triaged against two governance anchors: the AGENTS.md documentation gate (an in-repo documentation issue is required for user-facing changes before PR merge) and the org constitution's Cross-Repo Documentation rule (an unbound-force/website tracking issue is required for workflow/behavior changes). When the underlying change modifies a scaffolded command such as uf.unleash, a well-formed companion issue lists all three sync surfaces — the canonical .opencode/commands/<name>.md file, the internal/scaffold/assets/ embedded mirror, and the scaffold test pins (scaffold_test.go) — because TestEmbeddedAssets drift tests enforce byte-for-byte parity between canonical and embedded copies, so documentation verification is incomplete without confirming all three. Presence of these anchors, named surfaces, and binary-verifiable scope checkboxes (CHANGELOG reflects final behavior, docs cover the feature, website issue lands) supports a VALID/enhancement/objective classification without code inspection; the duplicate check must confirm no other open issue covers the same behavior (e.g., a different /uf.unleash behavior like the clarify ask gate is not a duplicate).
