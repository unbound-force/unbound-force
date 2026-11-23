---
tag: instruction-lint
author: jay-flowers
category: context
created_at: 2026-11-16T15:41:14Z
identity: instruction-lint-20261116T154114-jay-flowers
tier: draft
as_of: "unbound-force@c2dfd4b8f15f9ae106c599aef9a0418bf075dd93"
---

The large command files (.opencode/commands/uf.unleash.md ~21.4k, uf.address-feedback.md ~17.2k, uf.review-council.md ~16.6k, uf.triage-issue.md ~15.2k, uf.review-pr.md ~13.9k tokens) all exceed the instruction-lint IL-030 8000-token file budget and IL-031 20000-token load budget at baseline, so a --changed lint run flags them on every edit. The in-confinement review signal is the linter's delta.tokens (the change's marginal contribution), not the absolute overage: the project's declared pre-flight check is advisory and passes with these warnings, and the durable fix is DE-001 extraction work rather than per-edit trimming.
