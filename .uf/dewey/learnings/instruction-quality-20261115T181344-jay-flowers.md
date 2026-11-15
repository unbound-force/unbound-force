---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-15T18:13:44Z
identity: instruction-quality-20261115T181344-jay-flowers
tier: draft
as_of: "unbound-force@88deba6f4dd4024b88f4306e4f717b44f0dc37d7"
---

When command prose adds a recovery branch that reuses a sibling branch's protocol via "treat it as X" while excepting one clause ("rather than routing to ..."), the exception severs the imported protocol's escalation semantics: a strict reader cannot decide whether the imported branch's repeat-failure escalation still applies, and the excepted path can read as a standing rule. Partial protocol imports must state which clauses import — especially terminal/escalation clauses — or enumerate the imported obligations inline. Observed in /uf.run-pipeline-plan's timeout.kind "storm" bullet: "treat it as idle ... rather than routing to the blocked-item protocol" leaves a second storm failure undefined at the orchestrator layer — the exact unbounded-retry incident class (#1079) the branch was added to prevent; tracked as follow-up #1115.
