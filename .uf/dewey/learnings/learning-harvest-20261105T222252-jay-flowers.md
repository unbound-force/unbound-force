---
tag: learning-harvest
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:22:52Z
identity: learning-harvest-20261105T222252-jay-flowers
tier: draft
---

Before committing harvested Dewey learnings (.uf/dewey/learnings/*.md) as part of a change, verify each file's body block is a complete record: every stored learning must end with terminal punctuation, and a mid-word ending is the truncation signature. Five files in PR #955's council-iteration-1 batch (contract-testing-20261105T214916, contract-tests-20261105T214916, curator-20261105T214916, error-handling-20261105T214911, workflow-resume-gates-20261105T214909) end mid-word ('after i', 'when t', 'then subm', 'the push-fa', 'git push <re') — the harvest path stored child proposals whose text was already cut off, so each learning's remedy clause is missing while the file still presents as a valid draft-tier record. Distinguishing evidence: same-batch learnings of greater length end with complete sentences, ruling out a display or line-length cap. A completeness check at harvest time (or at commit-staging time) catches the truncation class; the same persistence layer also stores near-duplicates without dedupe, so one fix addresses both defect classes.
