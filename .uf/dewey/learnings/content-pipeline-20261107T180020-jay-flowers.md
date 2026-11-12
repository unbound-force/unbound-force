---
tag: content-pipeline
author: jay-flowers
category: pattern
created_at: 2026-11-07T18:00:20Z
identity: content-pipeline-20261107T180020-jay-flowers
tier: draft
---

Before flagging a missing blog issue for a change, check the unbound-force/website repo in addition to the current repo's blog-labeled issues. The constitution's cross-repo documentation gate files website issues that can carry both docs and blog labels: PR 957's stop-shipping-plugin-lock story is tracked as unbound-force/website#382 ('Publish the stop-shipping-plugin-lock story: two writers, one lock, and the lockless install', labels docs+blog), while the unbound-force/unbound-force repo itself has no blog-labeled issue covering it. A current-repo-only blog-issue search produces a false 'missing blog issue' finding; the website repo is the publication surface where story-framed issues live.
