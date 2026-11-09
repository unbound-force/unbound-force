---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:25Z
identity: content-pipeline-20261108T224025-jay-flowers
tier: draft
---

Dispatched review-council runs prohibit gh issue creation and external API access, while the Curator persona's mandate is to search and file docs/blog issues. Graceful degradation that preserves the mandate: state the verification gap in the finding, embed both the duplicate-search command and the complete gh issue create command (title, label, body) in the recommendation, and use the PR body's issue references plus the change's shipped Dewey learnings as the evidence layer to verify manually afterward.
