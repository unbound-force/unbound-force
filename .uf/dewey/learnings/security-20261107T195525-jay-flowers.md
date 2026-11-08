---
tag: security
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:55:25Z
identity: security-20261107T195525-jay-flowers
tier: draft
---

A syntactic guard for git push targets should document not only the classes it closes (`/`, `.`, `..`, `:`, leading `-`, whitespace) but also the class it deliberately leaves open: a bare separator-free token matching a cwd-relative directory that is itself a git repository still resolves as a local-path push target, delivering the verified HEAD commit there and executing that repository's receive hooks with the pusher's privileges. Without a doc-comment note distinguishing accepted residual from incomplete guard, later reviewers cannot tell the two apart. The class closes fully only with semantic verification — `git remote get-url <remote>` failing closed on any error — which trades parse-time INVALID_INPUT rejection for a push-time probe. PR 948 shipped the syntactic completion (RemoteNameSchema rejecting `.`/`..`, INVALID_INPUT table rows in both push tools, byte-identical scaffold mirror), accepting the bare-name residual because every consuming command pins `remote` to structured values (HEAD_REMOTE from `gh pr view`, origin default); a planted-repository precondition plus an unpinned remote is required to reach it.
