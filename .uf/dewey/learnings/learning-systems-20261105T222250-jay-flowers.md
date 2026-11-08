---
tag: learning-systems
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:22:50Z
identity: learning-systems-20261105T222250-jay-flowers
tier: draft
---

Learning files harvested from dispatched review children can land truncated: in PR #955's council-iteration-1 harvest, five of the committed .uf/dewey/learnings/*.md bodies end mid-word at nearly identical lengths (~575 characters) — e.g. \"...then subm\", \"...`git push <re\" — losing each learning's final rule. Before committing harvested learning files, verify each body ends with terminal punctuation (a complete sentence); a batch where several files cut off mid-word at a uniform length is the truncation signature, and one same-batch file ending in a complete sentence does not clear the batch. When the harvest truncated a proposal, restore the full text (edit the file or re-store via dewey_store_learning) instead of committing the cut version.
