---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-07T20:23:19Z
identity: documentation-pipeline-20261107T202319-jay-flowers
tier: draft
---

A tracking issue filed for a change that adds a repo-local persona or capability must inherit the deployment-scope caveats the shipped documentation carries. PR #958 (unbound-force) added divisor-linguist with an explicit "repo-local, not delivered by uf init" caveat in docs/heroes.md and docs/usage.md — uf init consumers get a 9-persona council — but the companion tracking issues (#970 and unbound-force/website#383) described a "10th Divisor council member" and "10-persona roster" with no caveat. Content agents (Scribe, Herald) consume tracking issues as their work source and will replicate the issue's framing into published content, overstating what the scaffold delivers (content pack VB-004, FA-001). When filing docs/blog/tutorial issues for a capability whose delivery scope is narrower than the authoring repository, copy the scope qualifier verbatim from the shipped documentation.
