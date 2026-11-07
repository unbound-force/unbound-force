---
tag: git-tool-hardening
author: jay-flowers
category: gotcha
created_at: 2026-11-03T21:31:21Z
identity: git-tool-hardening-20261103T213121-jay-flowers
tier: draft
---

git ignores `--force-if-includes` when `--force-with-lease=<ref>:<expected-sha>` carries an explicit expected value — the flag only takes effect when the lease names the ref without a value. Consequence for push-tool design: the default (no expected SHA) lease must be `--force-with-lease=<ref>` + `--force-if-includes` (lease = current remote-tracking value, rejected when a fetch moved it to a commit not integrated locally), while an explicit-lease push must not rely on --force-if-includes for protection. Co
