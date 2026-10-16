---
tag: fix-tool-args-schema-mismatch
author: jay-flowers
category: pattern
created_at: 2026-10-16T16:13:33Z
identity: fix-tool-args-schema-mismatch-20261016T161333-jay-flowers
tier: draft
---

The fix-tool-args-schema-mismatch change dogfooded its own fix mid-pipeline: the checkpoint and get_progress tools failed with CHECKPOINT_ERROR ("undefined is not an object (evaluating 'input.replace')") throughout Steps 0-8 because @opencode-ai/plugin 1.4.10 silently stripped their runtime arguments — the exact bug the change fixes. When the code review could not finalize (a separate review-dispatch defect), the run exited, the user restarted opencode, and the restart loaded the fixed zod-args plugin code: get_progress immediately returned ok and all five retroactive checkpoints recorded cleanly. The pipeline's own resumability machinery thereby became a live end-to-end acceptance test of the fix before the PR was ever created — the strongest possible confirmation that the SDK now recognizes and forwards the tool arguments. When a change fixes an infrastructure bug that its own workflow depends on, expect the workflow to break and then heal across the restart boundary, and treat the healed behavior as primary verification evidence.
