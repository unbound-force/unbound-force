---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-05T21:49:11Z
identity: error-handling-20261105T214911-jay-flowers
tier: draft
---

A divergence gate built on FETCH_HEAD inherits the fetch's failure mode. When `git fetch <remote> <ref>` fails (network, auth, stale URL), git leaves FETCH_HEAD holding the previous fetch's tip — including fetches performed implicitly by `gh pr checkout` — so a subsequent `git rev-list --left-right --count HEAD...FETCH_HEAD` silently reports divergence against stale state. A stale zero right-count reads as 'no divergence' and the flow proceeds to a push that then fails with a misdiagnosed cause (the push-failure handler reports network/protection, never the fetch that never happened). Every divergence gate that compares against FETCH_HEAD must prescribe its fetch's own failure contract — STOP with the specific error before the comparison — because the comparison is only valid against a fetch that succeeded. Found reviewing /uf.address-feedback Section 4.4 (PR #955), where the rewritten fetch line shipped without the failure contract the same increment added for its checkout and push commands.
