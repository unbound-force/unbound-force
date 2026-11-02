---
tag: issue-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-02T03:36:20Z
identity: issue-triage-20261102T033620-jay-flowers
tier: draft
---

Refactor issues that cite file structure (line counts, line ranges, named prompt blocks) as their problem statement can be accurate at filing and stale at triage time. Unbound-force issue #662 (extract the ~500-line review-pr subagent prompt into a dedicated agent file) matched the post-review-pr-subagent-split state of uf.review-pr.md (1,239 lines, inline Task-tool prompt), but the closed parent epic #659 later replaced that architecture with the review-dispatch plugin flow: the command now loads the pre-flight and review-context skills via the skill tool, passes their results as parameters to dispatch_review_runs (which builds child prompts from .opencode/agents/{agent}.md definitions), and contains zero subagent/Task-tool matches at 1,075 lines. Cheap staleness verification at triage: (1) grep the issue's architectural markers in the named file, (2) read the claimed line range, (3) check the parent epic's state. When the premise is superseded by completed work, close as duplicate-of-completed-work rather than re-scoping the issue's obsolete mechanism.
