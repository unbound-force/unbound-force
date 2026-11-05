---
tag: commit-scope
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:48:43Z
identity: commit-scope-20261105T174843-jay-flowers
tier: draft
---

Feature-branch diffs that bulk-add knowledge-capture files (dewey learnings, session notes) can be provenance-audited in a few samples: read the earliest-timestamped entries and check whether the body references a different PR/change than the branch's active change. In unbound-force PR #949's review, sampling showed .uf/dewey/learnings entries dated 2026-11-02 whose text documented PR #920's review and the #889 model-matrix scaffold deployment — learnings produced during those already-merged workflows, committed on the consumer-bounds branch. AGENTS.md's carve-out covers only learnings produced during the active change's workflow, so the created_at timestamp plus the referenced-PR pair is the discriminator between in-scope workflow output and out-of-scope catch-up files; apply the same test to small unexplained lockfile/config syncs riding the diff.
