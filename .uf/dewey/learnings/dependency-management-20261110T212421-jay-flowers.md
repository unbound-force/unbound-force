---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-10T21:24:21Z
identity: dependency-management-20261110T212421-jay-flowers
tier: draft
---

A documented manual verification command is a gated surface whenever it is the only verification path for a governed field. Precedent: unbound-force's uf doctor manifest-install check reads direct pins but never the npm `overrides` block, so the runbook's `(cd .opencode && npm ls zod toml)` command is the only consumer-facing way to confirm security overrides landed in a nested install tree — yet no test covers it, and when the effect 4.0.0-beta.103 override landed (PR adding GHSA-chp3-j5qr-2vfr remediation), every machine-enforced surface moved while the command, its expected-versions sentence, the section heading, and the pin table all stayed stale. Fix pattern, same shape as the repo's existing findProjectRoot drift tests: parse the doc's fenced verification command and assert its package list equals the keys of the manifest's machine-readable `overrides` block, so adding an override without extending the runbook fails the build. Rule: when a change grows a governed set, grep the runbook for every enumeration of that set — commands and headings read as examples but are procedure — and prefer gate-checking the doc enumeration against the machine-readable source over relying on the sweep.
