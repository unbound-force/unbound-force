---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:10:48Z
identity: ci-cd-20261107T161048-jay-flowers
tier: draft
---

When a lockfile removal is compensated by a new CI security scan added to the maintainer repository's workflow, scaffold consumers receive no equivalent: scaffolds deploy manifests and plugin sources, not workflow jobs. Pair the compensating job with a copyable recipe in the rollout guidance (the pinned npm, the --package-lock-only throwaway-lock generation, and the scanner invocation) so each consuming repository can reproduce the posture; a prose instruction to "review the scanner-loss posture first" leaves the recipe locked inside the maintainer's workflow file where consumers will not find it.
