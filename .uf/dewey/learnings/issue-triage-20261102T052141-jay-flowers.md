---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T05:21:41Z
identity: issue-triage-20261102T052141-jay-flowers
tier: draft
---

Issues that bundle a governance amendment with a code enhancement should be split at triage time. Constitution amendments follow their own PR/approval/migration-plan process and are exempt from the spec workflow, while the companion code enhancement requires the spec workflow and carries test gates; when bundled, the code path inherits the amendment's cross-repo alignment review timeline (hero constitution alignment within one release cycle) without needing it. Triage signal: an issue whose "Proposed actions" section contains both a CLI/scaffold code change and a constitution/wording change. The two typically have independent value and no implementation ordering dependency, so splitting loses nothing and lets each proceed through its own verification gate.
