---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:48:52Z
identity: determinism-20261116T014852-jay-flowers
tier: draft
---

instruction-lint --changed scopes only delta.files to the change; its findings array stays file-scoped. Reviewing the feedback-triage-artifact-tool change, IL-041 fired at .opencode/commands/uf.address-feedback.md:216 (Section 1.3's four gh api --paginate calls) even though the diff's changed hunks touched only the Progress Tracking note (lines 68-72) and Section 4.7 (lines 1334-1461). Cross-check every IL-040/IL-041/IL-042 hint's line against the diff's changed hunks before promoting it to a DE finding: the caller's changed-line confinement rule drops pre-existing prose even when the linter hints at it, and the determinism pack itself governs only new and modified instruction.
