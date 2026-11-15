---
tag: dependency-pins
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:11:47Z
identity: dependency-pins-20261110T211147-jay-flowers
tier: draft
---

Adding a security override to a pinned-dependency set is a multi-surface lockstep change: the canonical manifest overrides block (.opencode/package.json), the scaffold mirror (internal/scaffold/assets/opencode/package.json), the provenance record (plugin-package-provenance.json), the pinning tests (scaffold_test.go), the docs runbook (docs/osv-transitive-scan.md: table row, verification command, provenance enumeration, AND the rollback paragraph), and the CHANGELOG entry, which carries its own "Rollback: revert this commit to restore <previous version>" clause per the #957 stop-shipping-plugin-lock precedent. The rollback documentation must be per-override: PR #1031's first push left the runbook rollback sentence reading "Rollback is a revert of the PR #957 pin commit" — a sentence that cannot remove an override introduced later (reverting #957 would reintroduce the zod/toml advisories rather than restore effect 4.0.0-beta.83), drawing a HIGH gatekeeping finding because the Constitution's Gatekeeping Integrity rule makes a documented rollback path a conjunctive condition for any pinned-dependency update. The remediation (head commit b909c9b3) uses the resilient phrasing: "Rollback is a revert of the pin commit for the override in question" plus a per-PR restoration list — the pattern any future override addition should copy.
