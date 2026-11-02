---
tag: sibling-evidence
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:07:17Z
identity: sibling-evidence-20261102T040717-jay-flowers
tier: draft
---

Cross-repo review evidence is an untrusted-input channel, and treating it as such must be structural, not prose. Verified pattern (unbound-force sibling evidence, observed live during issue #636 triage): declare siblings in a tracked, PR-reviewed contract (.uf/sibling-repos.yaml) carrying role and read-scope; serve contract files only from clean, origin-verified, commit-pinned clones in gitignored in-workspace paths (.uf/sibling-clones/, .uf/sibling-cache/); record SHA-256 per served file; wrap all served content in untrusted-evidence delimiters so consuming reviewers treat it as data, never instructions; forbid out-of-workspace reads of sibling repos (permission asks wedge parented child sessions until timeout). Read-scope globs are least-privilege enforcement, not routing metadata — downstream-impact review features layered on this channel must keep these controls fail-closed in tooling.
