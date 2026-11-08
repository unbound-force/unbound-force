---
tag: command-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T20:23:08Z
identity: command-authoring-20261105T202308-jay-flowers
tier: draft
---

When a change adds a precondition rule for a tool call in an agent command file (PR #955's 'never call get_progress or checkpoint from another branch' in uf.address-feedback.md), enumerate every other invocation of that tool in the same file — especially early-running resume-time instructions like SESSION-RESUME guards and TodoWrite re-initialization, which execute before the main flow's setup steps. PR #955 updated the Progress Tracking section and added Section 1.2.1 but left two unconditional resume-path get_progress calls (lines 18 and 75), re-opening the wrong-branch-key failure (#954) through the resume path. Grep the tool name across the command and its byte-identical scaffold mirror, and extend the phrase-presence contract-test clauses to lock the corrected ordering.
