---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-10T23:31:57Z
identity: determinism-20261110T233157-jay-flowers
tier: draft
---

Fix-shape conversion gotcha (PR 1024, uf.triage-issue.md): when a security fix requires decomposing a single shell block (e.g., removing a quoted heredoc that composed gh api payloads), splitting it into numbered prose steps — bash mktemp/chmod/echo, host file-write to the printed path, bash gh api --input + rm -f — converts a DE-005-exempt one-call form into a DE-005 [MUST] violation, because every argument after the first call is the first call's printed output and the composed text is a pre-confirmed judgment product. The correct fix shape is extraction to an intent-level tool taking the confirmed text as a validated parameter (shipped precedent: file_feedback_issue and post_feedback_replies in .opencode/plugins/uf-workflow/index.ts, whose docstrings state that judgment products arrive as parameters and title/body reach gh as an argument-array element and a temp file, never shell text). Reviewers checking security-motivated rewrites of command templates should apply the DE-005 sequence test to the NEW call shape, not just the removed vulnerability — the pre-change single-bash-block form is exempt by one-call calibration, so the rewrite itself introduces the violation and the extraction belongs in the same change.
