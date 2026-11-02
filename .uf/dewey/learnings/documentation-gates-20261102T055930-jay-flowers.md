---
tag: documentation-gates
author: jay-flowers
category: gotcha
created_at: 2026-11-02T05:59:30Z
identity: documentation-gates-20261102T055930-jay-flowers
tier: draft
---

Documentation-gate tracking issues must carry the repo's docs label at filing time to stay visible to label-based triage. In unbound-force/unbound-force (2026-11-02, init-override-stubs-gitignore-sync review), doc-gate issue #916 was filed with correct title and scope and recorded in the change's notes.md, but carried no labels — while every other open docs issue in the repo carries the `docs` label (30/30 surfaced by `gh issue list --label docs --state open` at review time) and the model-matrix-delegation precedent #893 carried it. An unlabelled doc-gate issue is invisible to the label search that Curator duplicate-checks and docs triage rely on, risking duplicate filing and silently untracked documentation-gate work even though the gate's substance (issue exists, open, referenced in notes.md and the PR body) is satisfied. Fix at filing: include `--label docs` in the `gh issue create` command for doc-gate issues. Fix at review: when checking the documentation gate, verify the label alongside the issue's existence.
