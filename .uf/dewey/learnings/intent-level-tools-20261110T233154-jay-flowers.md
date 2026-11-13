---
tag: intent-level-tools
author: jay-flowers
category: pattern
created_at: 2026-11-10T23:31:54Z
identity: intent-level-tools-20261110T233154-jay-flowers
tier: draft
---

Splitting an atomic single-shell-block sequence (create temp file, write payload, post, cleanup) into multi-call prose choreography to remove a shell-injection vector trades the injection for an abort window: cleanup instructions living in the final call never execute when the run stops between calls, and transient paths must be agent-transcribed across calls. When decomposing shell blocks for injection safety, keep the sequence inside one owner -- an intent-level tool receiving the composed text as a validated parameter, with cleanup in a finally block. Shipped reference: .opencode/lib/feedback-replies.ts (PR #1007, learning security-20261110T151933).
