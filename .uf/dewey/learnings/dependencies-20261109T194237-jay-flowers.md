---
tag: dependencies
author: jay-flowers
category: gotcha
created_at: 2026-11-09T19:42:37Z
identity: dependencies-20261109T194237-jay-flowers
tier: draft
---

When documentation compensates for a verification-tool gap ("uf doctor does not read npm overrides, so confirm the nested copies manually"), scope the manual check to every item the same document classifies as security-relevant, not just the entries in the advisory-driven table. Observed on unbound-force PR #1000: docs/osv-transitive-scan.md's new pins section names three security overrides (toml, uuid, zod) in its provenance paragraph, but the adopter verification command (`npm ls zod toml`) and the expected-version sentence cover only zod and toml, leaving the uuid 13.0.1 override unconfirmed for anyone following the doc. The manifest (.opencode/package.json overrides block) and the provenance record (internal/scaffold/testdata/plugin-package-provenance.json security_overrides) both carry uuid as a security override, so the document and its own verification step disagree on scope.
