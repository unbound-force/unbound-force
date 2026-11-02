---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:10:13Z
identity: issue-triage-20261102T041013-jay-flowers
tier: draft
---

Cross-repo scaffold-sync sweeps filed as task issues in a meta/tooling repository (e.g., unbound-force/unbound-force#591 "update scaffolded files in complytime repositories") are valid tracking issues but cannot be implemented by a code agent dispatched against the meta repo — the executable work lives in each downstream repository. Established execution pattern: per-repo issues/PRs run `uf init` and adapt the diff (replicator #52/#53 for the uf.* command namespace rename; complytime-demos PR #80 and complyctl PR #849 for the complytime org), with the meta issue serving as the umbrella tracker. When triaging such issues: (1) mark VALID as a tracking artifact, (2) enumerate the target repositories in the assessment, (3) route implementation to per-repo execution rather than the meta repo's create-PR next step, and (4) check for open scaffolding bugs in the tool before endorsing the sweep — unfixed scaffold defects (org-framing leaks into external repos, stale pre-uf.* command references, schema installed without aligning consumer openspec/config.yaml) mean each `uf init` output requires human review-and-adapt per repo, not mechanical merge, so that review gate must stay in the sweep checklist.
