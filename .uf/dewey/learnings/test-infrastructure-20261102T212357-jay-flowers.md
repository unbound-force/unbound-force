---
tag: test-infrastructure
author: jay-flowers
category: gotcha
created_at: 2026-11-02T21:23:57Z
identity: test-infrastructure-20261102T212357-jay-flowers
tier: draft
---

In the unbound-force plugin test suite, the withBun helper (helpers/bun-runtime.ts) shims Bun.YAML.parse with JSON.parse so tests run under Node. Consequence: any feature whose production path parses YAML is only ever tested against the JSON subset of YAML — comments, unquoted keys, and block-list syntax never reach the production parser, and the deployed YAML template (the canonical day-one input) is never parsed by any test. In the preflight-declared change, all fixtures were JSON.stringify bodies or injected JSON.parse, so the 19-line comment-heavy .uf/preflight.yaml template shipped untested. Precedent for closing the gap: model-matrix-tracked.test.ts reads the deployed matrix text and validates it through parseYamlSubset (a comment-aware subset parser) plus schema validation. When reviewing or writing tests for a YAML-consuming feature in this repo, verify that (a) the deployed template text is parsed by at least one test, and (b) at least one hand-written YAML fixture with comments exercises the production parse path — JSON fixtures alone prove nothing about the grammar the shipped file uses.
