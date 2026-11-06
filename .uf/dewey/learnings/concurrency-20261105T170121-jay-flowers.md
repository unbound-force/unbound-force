---
tag: concurrency
author: jay-flowers
category: pattern
created_at: 2026-11-05T17:01:21Z
identity: concurrency-20261105T170121-jay-flowers
tier: draft
---

When several asynchronous abort sources (wall-clock timer, idle watchdog, parent cancellation) can end the same run through a shared abort-cause flag, every writer must set the flag conditionally (only when it is still null) or a later-firing timer can overwrite the classification of an earlier abort. In the unbound-force agent executor, the timeout callback assigns `abortCause = "timeout"` unconditionally; once an idle-timeout classification was added and consumed by kind-branching recovery logic, the ungu
