---
tag: dependency-pins
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:05:22Z
identity: dependency-pins-20261110T200522-jay-flowers
tier: draft
---

Adding or bumping a security override in .opencode/package.json requires five surfaces to move together: the canonical manifest, the embedded mirror internal/scaffold/assets/opencode/package.json, the scaffold test expectations (TestPluginPackageInputs_ExactVersions and TestPluginPackageInputs_ApprovedMaintenanceProvenance, including the verified_at date), the provenance JSON in internal/scaffold/testdata/plugin-package-provenance.json, and the docs/osv-transitive-scan.md runbook. The first four are machine-enforced (TestPluginPackageInputs_MirrorsMatchCanonical and the provenance drift test fail on any miss), but the runbook is not test-enforced: its section heading, pin table, and `npm ls` verification recipe silently go stale. PR #1026 (effect 4.0.0-beta.103 override) updated the runbook prose and provenance wording yet left the heading '(zod 4.6.6, toml 5.0.1)', the pin table, and 'npm ls zod toml' without effect — the doc surfaces are the ones tests cannot catch.
