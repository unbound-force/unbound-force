---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:49:52Z
identity: documentation-pipeline-20261110T194952-jay-flowers
tier: draft
---

The website-repo half of the documentation gate is missed recurrently for dependency-security bumps in unbound-force/unbound-force: PR #957 (zod/toml) initially lacked it until unbound-force/website#382 was filed to remediate, and PR #1024's effect override repeated the miss — in-repo #1034 was filed, but the website counterpart is absent even though #1034's body self-classifies the change as user-facing for hero-repo adopters. The in-repo issue's existence masks the cross-repo half; verification requires a separate search of unbound-force/website (keyword searches across open and closed states), because no in-repo artifact records the website counterpart when it was never filed.
