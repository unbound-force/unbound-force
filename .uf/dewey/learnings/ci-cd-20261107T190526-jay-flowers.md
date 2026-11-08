---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-07T19:05:26Z
identity: ci-cd-20261107T190526-jay-flowers
tier: draft
---

Paired-pin conventions carry two obligation classes with opposite failure modes. A pin coupled to a tool's declared version floor (GOTOOLCHAIN=go1.27.0 under a version-pinned `go install osv-scanner@v2.7.0`) fails CLOSED when the tool's floor advances — the install errors and CI goes red. A pin carrying a security-currency obligation ("bump on Go toolchain patch security releases") fails SILENT — after an upstream patch release the stale pin keeps CI green on an unpatched toolchain, and a keep-in-sync comment is not a tripwire. When writing a paired-pin comment, distinguish the two classes; for the currency class, either add a scheduled check comparing the pin against the latest patch release (e.g. folded into an existing ci_scheduled.yml job) or record the accepted staleness window at the decision point.
