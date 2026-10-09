---
tag: review-council-multi-model-fanout
author: jay-flowers
category: gotcha
created_at: 2026-10-02T16:48:40Z
identity: review-council-multi-model-fanout-20261002T164840-jay-flowers
tier: draft
---

Branch opsx/review-council-multi-model-fanout (2026-10-02): The review-council command's strict immutable-diff requirement blocked a local pre-PR review of uncommitted work because git rev-list --count main..HEAD was zero (all implementation lived only in the working tree), producing an empty base_sha...head_sha diff. The fix made the requirement contextual: PR reviews still resolve and use immutable base/head SHAs, but a local no-PR review where the branch head resolves to the same SHA as base falls back to the working tree (tracked edits via `git diff <base_sha> -- .` and untracked files via `git ls-files --others --exclude-standard`, each a full-content addition). Two hard constraints: the canonical `.opencode/commands/uf.review-council.md` and its embedded scaffold asset `internal/scaffold/assets/opencode/commands/uf.review-council.md` must be edited byte-identically (enforced by TestCommandContracts_CanonicalScaffoldParity), and several phrases are enforced by normalizeCommandText substring checks (lowercase + whitespace-collapse, backticks preserved) so wording like `base_sha...head_sha` must be preserved exactly.
