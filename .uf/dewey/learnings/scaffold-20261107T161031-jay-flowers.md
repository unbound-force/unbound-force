---
tag: scaffold
author: jay-flowers
category: gotcha
created_at: 2026-11-07T16:10:31Z
identity: scaffold-20261107T161031-jay-flowers
tier: draft
---

A lockless `npm install --omit=dev` still resolves devDependencies, so shipping a manifest without devDependencies only protects FRESH scaffolds from the npm 10.9.9 Arborist crash on the vitest 5 peer graph. A refresh path that preserves a target's existing devDependencies (stop-shipping-plugin-lock D2a, preservePluginDevDependencies) reintroduces the crash for every previously-scaffolded target and for the meta repo itself, while the version gate and remediation text still present npm 10 as supported. When a mitigation narrows a failure mode, enumerate the populations it does and does not cover (fresh vs previously-scaffolded targets) and record migration-path failures in the design's risk register.
