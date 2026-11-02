---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-10-31T02:51:32Z
identity: documentation-pipeline-20261031T025132-jay-flowers
tier: draft
---

Reviewer-side verification of documentation-gate claims: when a change's notes.md or proposal claims that docs/blog/tutorial issues were filed, those claims are untrusted until verified. Verify each claimed issue at review time with `gh issue view <N> --repo <owner>/<repo> --json number,title,state,labels` — a claimed-but-missing, closed, or unlabeled issue means the gate is unfilled (HIGH, REQUEST CHANGES). Before treating a missing issue as a gap, also run `gh issue list --label <label> --state open` to check for existing coverage; a matching open issue counts even if the notes did not cite it. In the review-dispatch-diff-manifest review (PR #882, 2026-10-30), both claimed issues (unbound-force/unbound-force#880 and unbound-force/website#349) verified OPEN with the docs label, converting would-be HIGH findings into a clean APPROVE WITH ADVISORIES.
