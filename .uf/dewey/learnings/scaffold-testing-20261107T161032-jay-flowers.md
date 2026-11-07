---
tag: scaffold-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:10:32Z
identity: scaffold-testing-20261107T161032-jay-flowers
tier: draft
---

When an embedded asset intentionally diverges from its canonical source (scaffold mirror = canonical minus a block, e.g. devDependencies), a generic drift-test remediation like "Run: cp source internal/scaffold/assets/path" becomes actively wrong — following it reintroduces the diverged block and breaks the design invariant. Derive the remediation text per asset alongside the expectedEmbedded transformation so failure messages teach the correct fix.
