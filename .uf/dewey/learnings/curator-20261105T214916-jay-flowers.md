---
tag: curator
author: jay-flowers
category: pattern
created_at: 2026-11-05T21:49:16Z
identity: curator-20261105T214916-jay-flowers
tier: draft
---

divisor-curator dispatched as a review-council child cannot exercise its gh issue mandate: the dispatch prohibitions forbid creating issues and accessing external services, and the agent definition's `gh issue create` "ask" permission would wedge a parented child session (no one can answer the prompt). Working pattern (PR #955 pre-push review): run the Dewey prior-learnings queries first (Dewey MCP is in-context and therefore permitted; if the embedding model is down, skip Step 0 informationally), then submit the documentation-gap finding with the fully drafted issue title and body plus the exact `gh issue list` duplicate-search and `gh issue create` commands in the recommendation so the parent session or developer files manually — source the repo value from AGENTS.md or omit `--repo` (targets the current checkout), since `gh repo view` detection is equally prohibited. State in the finding that the duplicate-issue search could not run, so an existing issue moots the finding.
