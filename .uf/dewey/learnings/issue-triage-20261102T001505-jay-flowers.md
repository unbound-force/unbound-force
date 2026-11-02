---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:05Z
identity: issue-triage-20261102T001505-jay-flowers
tier: draft
---

Phased-migration issues that bundle post-window work cannot close when the phase's primary deliverable completes. In a cutover phase that cuts a support branch archived after a 90-day window, the deferred removal of v1 implementations ("post-window: remove the v1 server() implementations once the support window closes and no v1 consumers remain") belongs in its own follow-up issue linked to the umbrella, carrying its own verification condition (no remaining v1 consumers), so the phase issue closes at cutover completion — gate, conversion, release tag, and support cut — instead of staying open for the length of the support window.
