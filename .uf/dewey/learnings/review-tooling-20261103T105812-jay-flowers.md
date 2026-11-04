---
tag: review-tooling
author: jay-flowers
category: pattern
created_at: 2026-11-03T10:58:12Z
identity: review-tooling-20261103T105812-jay-flowers
tier: draft
---

When a PR edits files that have byte-identical scaffold twins (e.g. `.opencode/plugins/<name>/index.ts` and `internal/scaffold/assets/opencode/plugins/<name>/index.ts`), reviewers can verify twin parity deterministically from the diff headers alone: each file section of a git diff carries an `index <old-blob>..<new-blob>` line, and matching new-blob hashes across both file sections prove byte-identity without reading or diffing the file contents a second time. Applied to unbound-force PR #927 (fix-uf-finale-ci-poll-gh-schema): `.opencode/plugins/uf-finale/index.ts` and its scaffold twin both showed `index a3591437..74b97f23`, and `.opencode/plugins/review-dispatch/index.ts` plus its twin both showed `index 2b78ac4b..8f26f5d7` — parity confirmed with zero extra chunk retrievals. Caveats: the proof holds when both twins are modified in the same diff; if only one twin of a known pair appears in the changed-file list, parity is broken and is a review finding (the scaffold drift tests would also fail in CI).
