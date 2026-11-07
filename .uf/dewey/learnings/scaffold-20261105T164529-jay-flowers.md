---
tag: scaffold
author: jay-flowers
category: pattern
created_at: 2026-11-05T16:45:29Z
identity: scaffold-20261105T164529-jay-flowers
tier: draft
---

When a spec scenario's outcome is a negative composition ("the next run neither overwrites nor reads it"), piecewise mechanism assertions leave a traceability gap: predicate tables and command-string pins verify each mechanism in isolation, but the composed observable behavior is never exercised, so regressions through adjacent paths (cleanup passes, write-list reintroduction, install-mechanism rewrites of the string pin) evade the suite. Close the gap with one scenario-level test that pre-seeds the foreign artifact before Run() and asserts byte-identical survival afterward. Gotcha for the uf scaffold seam specifically: reviewPluginRunOpts fakes npm, so a composed Run()-level test verifies the "does not overwrite" half end-to-end but the "does not read" half is observable only through the install command-string pin — the scenario test must carry both assertions (file bytes unchanged + command carries --no-package-lock) to pin the full AND clause. Precedent: PR 957 stop-shipping-plugin-lock, SC-FR-003 "Lock written by another tool" scenario at specs/scaffold/spec.md:53.
