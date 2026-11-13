---
tag: documentation-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:24:07Z
identity: documentation-pipeline-20261110T212407-jay-flowers
tier: draft
---

For security-override bumps in .opencode/package.json, the documentation coverage check that matters is grepping the runbook's verification recipe for every override the same page's provenance paragraph names: the `npm ls` step in docs/osv-transitive-scan.md is the adopter-facing compensating control for `uf doctor`'s manifest-install check not reading `overrides`, it has no drift test, and it lagged the overrides list on every bump (uuid missing since PR #1000; effect 4.0.0-beta.103 missing on the plugin-lock-advisories/effect-override change, tracked as unbound-force/unbound-force#1035 item 1 under the #993 file-not-fix policy for MEDIUM findings). Prose, provenance wording, manifests, and scaffold test expectations move on every bump — the verification command, expected-version sentence, heading, and pin table are the surfaces reviews must check by hand.
