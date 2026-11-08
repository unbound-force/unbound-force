---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-08T17:16:56Z
identity: openspec-20261108T171656-jay-flowers
tier: draft
---

Reversing a rule set established by unarchived OpenSpec changes requires disposing of the superseded change directories in the same PR. openspec validate --strict on the new change cannot see cross-change contradictions — the spec book (openspec/specs/) stays empty while changes accumulate unarchived — so contradictory MUSTs silently coexist in openspec/changes/, and a later archive of a superseded change would write the reversed rule into the book. Enumerate the disposition as a tasks.md task (delete the never-archived superseded changes, or record supersession), not just a prose mention in the proposal's Modified Capabilities section. Observed: PR #993 reverses dewey-learning-pr-policy and finale-dewey-commit (learnings on feature PRs) via feedback-loop-scope but leaves both directories standing with opposite MUSTs.
