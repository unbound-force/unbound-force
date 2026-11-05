---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-05T21:49:11Z
identity: error-handling-20261105T214911-jay-flowers
tier: draft
---

A divergence gate built on FETCH_HEAD inherits the fetch's failure mode. When `git fetch <remote> <ref>` fails (network, auth, stale URL), git leaves FETCH_HEAD holding the previous fetch's tip — including fetches performed implicitly by `gh pr checkout` — so a subsequent `git rev-list --left-right --count HEAD...FETCH_HEAD` silently reports divergence against stale state. A stale zero right-count reads as 'no divergence' and the flow proceeds to a push that then fails with a misdiagnosed cause (the push-fa
