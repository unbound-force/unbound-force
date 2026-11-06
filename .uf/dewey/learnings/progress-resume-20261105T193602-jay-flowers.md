---
tag: progress-resume
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:36:02Z
identity: progress-resume-20261105T193602-jay-flowers
tier: draft
---

Checkpoint-resume commands that freeze immutable context (e.g., head_sha fetched from GitHub in /uf.address-feedback Section 1.2) must not gate resume on exact equality with the frozen value. The command's own failure path — push failure with "Local commits preserved" (Section 4.4) — leaves local HEAD a strict descendant of the frozen head_sha, so an exact HEAD == head_sha check STOPs the retry run before get_progress runs, making the documented "resume from the first incomplete phase" unreachable. Gate on ancestry instead: on the head branch, accept HEAD when `git merge-base --is-ancestor <head_sha> HEAD` holds (ahead = resumable); STOP only on divergence or behind-remote. When auditing a new STOP condition, check it against every state the command's own failure paths can leave behind, not just hostile states. And when correcting command prose in this repo, remember it ships in four synchronized places — the canonical .opencode/commands/ file, the byte-identical internal/scaffold/assets mirror (drift-enforced), the command_contract_test.go clauses, and the CHANGELOG entry — all four must change together.
