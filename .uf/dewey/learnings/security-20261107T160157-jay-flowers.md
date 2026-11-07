---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:01:57Z
identity: security-20261107T160157-jay-flowers
tier: draft
---

Git resolves a push target that is not a configured remote as a URL or path, and scheme-less separator-free tokens are still local paths: `.` is the current repository and `..` the parent directory. Git's own remote-name validation (check_refname rules) rejects `.`/`..` as remote names, so rejecting them excludes no legitimate remote. A RemoteNameSchema derived from a ref-name schema plus only a `/` rejection (the security-20261105T163851 fix pattern, shipped in PR 948) leaves this same local-path class open: `git_push` with `remote: ".."` pushes the verified HEAD commit to the parent-directory repository with no lease, and git-receive-pack executes that repository's pre/post-receive hooks with the pusher's privileges (destination must pre-exist as a repository). When a validation schema's doc comment cites an upstream tool's name-validation rules as its model, enumerate the full rule set, not only the separator that motivated the fix: complete the remote guard by rejecting `.`/`..` (or validating against git's remote-name rules), add INVALID_INPUT test rows for both values, and sync the byte-identical scaffold mirror.
