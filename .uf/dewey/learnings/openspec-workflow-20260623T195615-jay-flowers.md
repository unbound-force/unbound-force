---
tag: openspec-workflow
author: jay-flowers
category: pattern
created_at: 2026-06-23T19:56:15Z
identity: openspec-workflow-20260623T195615-jay-flowers
tier: draft
---

When extending an existing OpenSpec change with new tasks that supersede completed tasks (e.g., task 5.2 implemented the old `AI-assisted-by: /finale` format, then tasks 6.1-6.2 replaced it with `Assisted-by: <model>`), the review council will flag the contradiction between the completed task's description and the spec's current requirements. The solution is to annotate the completed task with a supersession note (e.g., "superseded by tasks 6.1-6.2") rather than unchecking it, preserving the audit trail while making the phased approach explicit. Also, if a scaffold sync task (like 8.1) was marked complete but later tasks modify the same file, uncheck the sync task to prevent stale scaffold copies.
