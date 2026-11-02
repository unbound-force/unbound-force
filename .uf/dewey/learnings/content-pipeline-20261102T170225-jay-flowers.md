---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-02T17:02:25Z
identity: content-pipeline-20261102T170225-jay-flowers
tier: draft
---

Documentation-gate tracking issues must carry the repo's conventional `docs` label to stay discoverable by the content pipeline's label-based duplicate checks. Verified during the split-host-resolution-failure-classes review (2026-11-02): the change's in-repo issue #919 and website issue #370 were filed with accurate bodies and correct CHANGELOG cross-references but zero labels, while all four precedent issues the change's own tasks.md cited (#778, #891, website#316, website#360) carry the `docs` label, and both repos maintain large label-scoped landscapes (36 open docs-labeled issues in unbound-force/unbound-force, 50 in unbound-force/website). Full-text search (`gh issue list --search "<term>"`) finds the unlabeled issues, but the label filter (`--label docs`) does not. Curator verification of a documentation disposition should therefore check `gh issue view <n> --json labels` against the cited precedents, not just issue existence; the filing-side fix is `gh issue edit <n> --add-label docs` on both repos before PR merge.
