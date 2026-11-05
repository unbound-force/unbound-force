---
tag: learnings-harvest
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:22:53Z
identity: learnings-harvest-20261105T222253-jay-flowers
tier: draft
---

When a review council's lesson-harvest path writes .uf/dewey/learnings/*.md files, verify each committed body ends with a complete sentence before pushing. On PR #955 (branch quick/address-feedback-pr-checkout, commit 64457c88), five of six council-iteration-1 learnings were committed truncated mid-word ('...after i', '...then subm', '...exactly when t', '...(the push-fa', '...git push <re') while all 30 feedback-assessment learnings from the same PR were complete — the defect clustered in the child-proposal harvest batch, and no gate validates learning-file completeness. Detection is cheap: read the file and confirm terminal punctuation (the git diff hunks show the same truncation). Full texts survive only in the dispatch artifact; restore from there.
