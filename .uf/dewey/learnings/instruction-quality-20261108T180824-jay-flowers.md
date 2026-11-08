---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-08T18:08:24Z
identity: instruction-quality-20261108T180824-jay-flowers
tier: draft
---

When replacing a command inside a fenced Markdown code block, verify fence parity after the edit: adding a new closing fence while leaving the original produces a stray fence that opens a second code block and silently absorbs the following prose and examples until the next bare fence (an info-string fence like ```json cannot close it, because closing fences carry no info string). Instruction linters do not check fence balance, so the defect survives a clean lint run and only manifests for consumers that parse rendered Markdown structure. Observed on unbound-force PR #959: divisor-linguist.md Step 1 gained a stray fence that swallowed the wrapper-invocation rule, exit-code semantics, and the linter JSON example.
