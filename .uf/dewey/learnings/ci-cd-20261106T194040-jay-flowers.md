---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-06T19:40:40Z
identity: ci-cd-20261106T194040-jay-flowers
tier: draft
---

An exact tool pin duplicated across a Makefile variable and a CI workflow step, twinned only by "keep in sync" comments, has no drift guard: bump one surface and the other fails or silently diverges. Convert the comment protocol into an enforced gate using the repo's drift-test pattern — a lightweight test asserting the pin string appears identically in both surfaces (e.g., PLUGIN_TEST_NPM in the Makefile and the osv_plugin_transitive step in ci_security.yml both carrying npx -y npm@11.6.2 in unbound-force PR #957).
