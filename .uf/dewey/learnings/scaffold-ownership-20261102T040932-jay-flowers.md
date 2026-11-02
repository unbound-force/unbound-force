---
tag: scaffold-ownership
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:09:32Z
identity: scaffold-ownership-20261102T040932-jay-flowers
tier: draft
---

When a scaffold engine with skip-if-present semantics adds a new never-overwrite file-protection class, every output message and flag help text that assumes --force is the universal overwrite path becomes actively false for that class. In unbound-force, the shield-constitution-from-force change (#495) added the isNeverOverwrite guard for .specify/memory/constitution.md but routed its skips into the existing undifferentiated Result.Skipped bucket, which printSummary renders with a single generic "(use --force to overwrite)" hint — so the tool advises an action that cannot succeed, and the --force help text ("Overwrite all existing files") overstates the flag. The Result model conflates three distinct skip reasons (never-overwrite protected, tool-owned with identical content, user-owned) behind one bucket and one hint. Pattern: introducing a protection class requires a concurrent audit of the skip-reason taxonomy, the summary message templates, and the flag help text; a spec scenario mandating placement in Result.Skipped (SC-001 of shield-constitution-from-force) constrains the fix to reason-annotation or a parallel bucket rather than removal from Skipped.
