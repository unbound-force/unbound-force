---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-04T14:24:20Z
identity: openspec-20261104T142420-jay-flowers
tier: draft
---

Bundled issue fixes on a spec-governed branch must be recorded in the governing OpenSpec change before merge. When extra fixes land as separate commits on an opsx/* branch, the change directory — not the commit log or the issue tracker — is the audit surface: each bundled fix needs a What Changes bullet (or a "Bundled fixes" note) naming its issue, a tasks.md task, and a "(Spec: <change-path>, Fixes: #N)" tail on its CHANGELOG entry, matching the convention the change's own documentation task prescribed. PR 949 (review-dispatch-consumer-bounds) shipped #946 and #942 unrecorded while #941/#945/#943 were fully recorded; the asymmetry between recorded and unrecorded bundled work is what exposes the gap at review time.
