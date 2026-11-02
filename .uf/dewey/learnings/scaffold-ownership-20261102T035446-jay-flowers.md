---
tag: scaffold-ownership
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:54:46Z
identity: scaffold-ownership-20261102T035446-jay-flowers
tier: draft
---

In the uf scaffold engine, isToolOwned=false (user/project-owned) grants NO protection from --force: the force path (internal/scaffold/scaffold.go, opts.Force branch) overwrites every asset except isNeverOverwrite entries, which is currently only the constitution. In-code comments or warning text claiming a file is preserved "without --force replacing" it, or that "--force re-scaffolds tool-owned files only" (both appear in scaffold.go — the uf/review-matrix.yaml branch comment and the stale-command-reference warning), must be verified against isNeverOverwrite, not isToolOwned; the accurate contract lives in docs/configuration.md ("survives plain uf init re-runs, reset to canonical only by --force"). Verified during unbound-force issue #649 triage. Corollary: project-owned classification of a schema-versioned contract file (uf/review-matrix.yaml, consumed by review-dispatch at dispatch time) needs a staleness signal — the user-owned skip path records only a Skipped disposition with a generic "--force to overwrite" hint and performs no canonical-vs-disk comparison, so seeded repositories accumulate stale matrix schema versions (v2 vs v3) with no drift indication until dispatch degrades.
