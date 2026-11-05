---
tag: intent-drift
author: jay-flowers
category: pattern
created_at: 2026-11-05T20:23:10Z
identity: intent-drift-20261105T202310-jay-flowers
tier: draft
---

When a command-file change adds an ordering or keying invariant for a named tool call (e.g., "never call get_progress or checkpoint from another branch"), enumerate every existing call site of that tool in the affected file before finishing the change: guard preambles, progress-tracking sections, TodoWrite resume text, and phase steps — plus the byte-identical scaffold mirror. The sweep is mechanically discoverable by grepping the tool name, which is cheaper and more reliable than reasoning about which sections "state the model". PR #955 repeated the PR #745 guard-block pattern (stored learning intent-drift-20261017T124835) despite that learning being on file: the new Progress Tracking section and Section 1.2.1 both asserted the branch-keying invariant while the SESSION-RESUME GUARD at the top of the same file still instructed an unconditional pre-checkout get_progress. A tool-name grep at authoring or review time catches the stale call site directly, before a reviewer has to.
