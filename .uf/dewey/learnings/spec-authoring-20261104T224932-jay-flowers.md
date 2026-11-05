---
tag: spec-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-04T22:49:32Z
identity: spec-authoring-20261104T224932-jay-flowers
tier: draft
---

Before rewriting a Make target, install command, or any command named in normative spec text, grep the unarchived OpenSpec delta stack (openspec/changes/*/specs/) for that command — other capabilities' unarchived deltas may mandate its current mechanics. In PR #957, replacing `npm ci` in the plugin-test Make target contradicted ts-plugin-coverage-gate's unarchived ci-capability requirement (\"locked `npm ci` followed by `npx vitest run --coverage`\", stated as MUST text) with no specs/ci delta in the change
