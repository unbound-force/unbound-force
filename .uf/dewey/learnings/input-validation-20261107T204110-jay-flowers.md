---
tag: input-validation
author: jay-flowers
category: pattern
created_at: 2026-11-07T20:41:10Z
identity: input-validation-20261107T204110-jay-flowers
tier: draft
---

When a tool-argument validation guard is deliberately syntactic — rejecting a value's shape but not its resolution semantics — record the accepted residual and its compensating control in the guard's own doc comment, not only in the review thread. Verified on unbound-force PR #948 round 5: RemoteNameSchema rejects '/', '.', '..', ':', whitespace, and leading '-', yet a bare separator-free token naming a cwd-relative directory that is itself a git repository still resolves as a local-path push target (verified HEAD commit delivered there, receive hooks executed with the pusher's privileges); the residual is accepted because all consuming commands pin 'remote' to structured values, with 'git remote get-url' as optional semantic verification. Residual decisions that live only in review threads do not survive context compression, get re-flagged in later rounds (round 4 flagged the missing '.'/'..' rejection; round 5 the bare-token residual), and leave future maintainers unable to distinguish an incomplete guard from a knowingly accepted boundary. The note's remediation surface mirrors the tightened-schema learning (git-tooling-20261107T165845): canonical schema file, byte-identical scaffold mirror under internal/scaffold/assets/opencode/ (drift tests enforce identity), and any artifact that restates the rule.
