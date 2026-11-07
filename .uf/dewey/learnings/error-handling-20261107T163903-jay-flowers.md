---
tag: error-handling
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:39:03Z
identity: error-handling-20261107T163903-jay-flowers
tier: draft
---

Runtime acceptance ranges, printed remediation strings, and the maintained escape hatch form one consumer-visible surface that must move together. In PR 957, uf doctor accepted npm 10-11 while activation crashed on npm 10.9.9 for devDependencies-carrying targets, the remediation constants recommended the crashing runtime (npm 10) and re-ran the crashing command, and the working workaround (npx -y npm@11.6.2) existed only in the source repo's Makefile, CI workflow, and AGENTS.md — surfaces never deployed to scaffolded targets. Any escape hatch the source repository relies on must be propagated into the remediation strings and version-gate messaging that ship to consumers, or the acceptance gate must exclude the runtime the escape hatch exists to avoid.
