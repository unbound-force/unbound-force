---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:09:16Z
identity: test-quality-20261106T180916-jay-flowers
tier: draft
---

When a drift test's expectation becomes a derived value (e.g., embedded asset = canonical source minus a block, such as the scaffolded .opencode/package.json being the repo manifest without devDependencies) rather than a byte copy, the failure remediation must describe the derivation. A stale "cp canonical → assets" hint in the error message reproduces the failure instead of fixing it, because copying the canonical file now violates the derived expectation. Rule of thumb: every time an expectedEmbedded-style transformation is introduced, audit the paired remediation text in the same change.
