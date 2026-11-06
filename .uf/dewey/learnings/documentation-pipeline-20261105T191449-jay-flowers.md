---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:14:49Z
identity: documentation-pipeline-20261105T191449-jay-flowers
tier: draft
---

A docs-gate task requiring issues in two repos must be verified per repo and per label, not by content coverage alone. PR #957: the current-repo half was correct (docs-labeled #964) but the website half was satisfied by a blog-labeled narrative issue that did not track the required documentation areas — a publication narrative is not a docs tracker. Verify each claimed gate artifact against the gate's actual requirement (repo, label, coverage).
