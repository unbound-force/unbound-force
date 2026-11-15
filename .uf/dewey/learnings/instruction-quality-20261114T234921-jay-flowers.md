---
tag: instruction-quality
author: jay-flowers
category: pattern
created_at: 2026-11-14T23:49:21Z
identity: instruction-quality-20261114T234921-jay-flowers
tier: draft
---

Linguist spec-review scope calibration: an OpenSpec change that proposes only TypeScript plugin/library source edits (.opencode/lib/*.ts, .opencode/plugins/*/index.ts), byte-identical .ts scaffold mirrors under internal/scaffold/assets/, tests, and .eval/ fixtures proposes no instruction files — divisor-linguist Spec Review Mode produces zero findings and APPROVEs. The instruction-file path table covers .opencode/agents|commands|skills|uf/packs .md files, AGENTS.md, and .md scaffold mirrors only; .eval/fixtures/&lt;name&gt;/guidance.md briefs and .ts mirrors fall outside it. Verified on the review-dispatch-a-reviewer-with-empty-ou change (#1052 empty-reviewer fail-closed fix), whose entire planned surface is executor/dispatch TS code, tests, and an eval fixture.
