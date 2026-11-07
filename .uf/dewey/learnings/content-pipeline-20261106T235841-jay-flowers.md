---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-06T23:58:41Z
identity: content-pipeline-20261106T235841-jay-flowers
tier: draft
---

Label-filtered duplicate-issue checks miss issues that carry alternative or missing labels: in unbound-force/unbound-force, a --label docs search returned 44 open docs issues but missed #977 (label "documentation") and #983 (no labels), both directly governing the change under review; only a keyword search (--search "lock") surfaced them. The label namespace is known-unstandardized (issue #825). Content-pipeline duplicate checks MUST pair every label-filtered search with a keyword search across open issues before concluding a docs/blog/tutorial issue is missing.
