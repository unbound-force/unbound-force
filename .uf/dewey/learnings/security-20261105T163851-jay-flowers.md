---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-05T16:38:51Z
identity: security-20261105T163851-jay-flowers
tier: draft
---

Git push targets resolve by fallback: a token that is not a configured remote name is treated as a URL, and a `/`-bearing scheme-less URL is a local filesystem path — so a `remote` tool parameter validated by a ref-name schema that permits `/` (rejecting only leading `-`, whitespace, `:`) accepts local-path push targets. `git push /path` copies the branch's commits to that repository, and git-receive-pack executes the destination repository's pre/post-receive hooks with the pusher's privileges (the destination must already exist as a repository; push does not create it). Remote-name parameters need a dedicated schema that also rejects `/`; branch/refspec parameters legitimately contain `/` (feat/x) and must not share that schema. Precedent: PR 948 shared PushRefNameSchema across branch and remote on git_push/git_push_force_with_lease; fix pattern: a RemoteNameSchema applied to both push tools plus INVALID_INPUT table rows (remote: "/tmp/repo") in the TS suite, landed in both the plugin and its byte-identical scaffold mirror.
