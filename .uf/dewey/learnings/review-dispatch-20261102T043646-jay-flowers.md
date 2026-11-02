---
tag: review-dispatch
author: jay-flowers
category: context
created_at: 2026-11-02T04:36:46Z
identity: review-dispatch-20261102T043646-jay-flowers
tier: draft
---

In unbound-force, the pre-PR review council is blind to originating-issue acceptance criteria by construction. The review-context skill's Protocol 2 (Issue Linking) runs only when a PR body is available — the consumer table marks it "conditional (explicit PR number only)" for /uf.review-council, and its error handling is advisory ("continue without blocking"), so it never gates a verdict. The local pre-PR path (/uf.unleash Step 6 Code Review) has no PR, so Protocol 2 is skipped entirely. Meanwhile the issue number recorded by tier routing (branch-scoped checkpoint state) and .openspec.yaml originating_issue are consumed only by /uf.finale to emit "Closes #N" in the PR body — review dispatch's build_review_prompts carries issue_content in triage mode only. Net effect: the council answers "is this code good?" with spec artifacts as the sole intent source, and the originating issue's ACs reach review only post-PR via /uf.review-pr. Verified against .opencode/skills/review-context/SKILL.md while triaging issue #563.
