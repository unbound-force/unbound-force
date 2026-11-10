---
tag: openspec
author: jay-flowers
category: gotcha
created_at: 2026-11-10T15:19:53Z
identity: openspec-20261110T151953-jay-flowers
tier: draft
---

Superseding a clause of an unarchived OpenSpec change leaves the supersession record only in the superseding change's proposal.md, which does not survive into openspec/specs/. A MODIFIED delta cannot target a requirement that is not yet canonical, so when the earlier change later archives, its requirement lands verbatim as a canonical MUST that contradicts shipped behavior, and nothing in the archive path surfaces the note. Record the supersession where the archive step will see it: a header note in the superseded change's spec naming the superseding requirement, or a tracking issue to reconcile at archive time. Precedent: feedback-reply-tools (PR #1007) superseding AFS-FR-002's reject+prefix recording clause in feedback-loop-scope via proposal note only.
