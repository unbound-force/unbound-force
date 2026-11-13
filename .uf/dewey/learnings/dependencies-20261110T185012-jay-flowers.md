---
tag: dependencies
author: jay-flowers
category: pattern
created_at: 2026-11-10T18:50:12Z
identity: dependencies-20261110T185012-jay-flowers
tier: draft
---

For a CVE in a nested transitive dependency that a direct dependency pins exactly (e.g. effect 4.0.0-beta.83 under @opencode-ai/plugin 1.19.0), bumping a root pin leaves the vulnerable nested copy in place; the npm manifest `overrides` block forces every lock resolution to the fixed version (zod precedent, reapplied for effect 4.0.0-beta.103 / GHSA-chp3-j5qr-2vfr). The lockstep set for this repo: canonical `.opencode/package.json` + scaffold mirror (D2a: identical minus devDependencies) + TestPluginPackageInputs_ExactVersions expected overrides + plugin-package-provenance.json (upstream git head, registry integrity sha512, verified date) + CHANGELOG entry documenting the gatekeeping conditions (advisory, verified SHA, rollback, validation evidence).
