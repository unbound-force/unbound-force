---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:08:00Z
identity: dependency-management-20261110T190800-jay-flowers
tier: draft
---

Security-override pin changes in unbound-force reliably sweep the test-enforced surfaces (both package.json manifests, the provenance testdata, the scaffold expectedOverrides maps, CHANGELOG) but the runbook's procedural surfaces lag exactly like narrative enumerations do: docs/osv-transitive-scan.md carries three further override enumerations that no gate covers — the override-verification command ("(cd .opencode && npm ls zod toml)"), its expected-version sentence ("Every zod entry should read 4.6.6 and every toml entry 5.0.1"), and the section heading ("Plugin dependency pins (zod 4.6.6, toml 5.0.1)"). When the effect 4.0.0-beta.103 override landed (PR #1026, GHSA-chp3-j5qr-2vfr), all three went un-swept, leaving the doc's only override-verification path unable to confirm the fix it had just documented (uf doctor's manifest-install check reads direct pins, not overrides). When adding or bumping an override, grep the runbook for every enumeration of the existing override set — fenced verification commands and headings included, not just prose lists — and extend each to the new package. Recurrence of the dependency-management-20261106T194037 lag pattern in procedural doc surfaces.
