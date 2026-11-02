---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:13:19Z
identity: issue-triage-20261102T001319-jay-flowers
tier: draft
---

In Unbound Force repos, scaffolded user-facing command files (e.g., .opencode/commands/uf.unleash.md) have byte-pinned embedded mirrors under internal/scaffold/assets/ enforced by scaffold drift tests, plus scaffold test pins in scaffold_test.go. When triaging documentation issues that track changes to scaffolded surfaces, verify the issue enumerates BOTH copies and the test pins — a docs issue naming only the .opencode/ path describes a fix that cannot land cleanly, because updating one side of a mirrored pair fails drift detection. Issue #819 models the complete pattern: it lists .opencode/commands/uf.unleash.md, internal/scaffold/assets/opencode/commands/uf.unleash.md, and internal/scaffold/scaffold_test.go as affected surfaces. This mirror-pair pattern recurs across the org (unbound-force uf init, gaze init, replicator init all carry drift-detection mirror pairs).
