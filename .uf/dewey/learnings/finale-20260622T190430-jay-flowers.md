---
tag: finale
author: jay-flowers
category: gotcha
created_at: 2026-06-22T19:04:30Z
identity: finale-20260622T190430-jay-flowers
tier: draft
---

When modifying slash command files that have scaffold asset copies (e.g., .opencode/commands/finale.md has a copy at internal/scaffold/assets/opencode/commands/finale.md), always verify byte-identity after changes by running the specific drift detection test: `go test -race -count=1 -run TestEmbeddedAssets_MatchSource ./internal/scaffold/`. The full scaffold test suite includes integration tests (TestRun_SchemaDistribution) that can timeout due to external dependencies (Dewey/Ollama), so targeting the specific drift test is more reliable for verification.
