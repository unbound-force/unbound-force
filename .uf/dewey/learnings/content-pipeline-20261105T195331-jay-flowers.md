---
tag: content-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:53:31Z
identity: content-pipeline-20261105T195331-jay-flowers
tier: draft
---

Curator blog-opportunity duplicate checks for the unbound-force org must span both repositories, not just the detected current repo: the blog tracking issue for a change may be filed in unbound-force/website rather than the current repo, and may carry combined labels (website issue #382 for stop-shipping-plugin-lock holds docs, blog together, framing the story "two writers, one lock, and the lockless install"). A blog-issue search scoped to the current repo alone would falsely report a missing blog issue.
