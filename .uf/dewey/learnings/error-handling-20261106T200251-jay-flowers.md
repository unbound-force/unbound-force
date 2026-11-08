---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-06T20:02:51Z
identity: error-handling-20261106T200251-jay-flowers
tier: draft
---

Advisory checks that deliberately fail open on tool errors should still report the failure cause rather than a guessed one. In internal/doctor/checks.go, checkTrackedPluginLock maps every `git ls-files` error to the message "not a git repository" and returns Pass — correct as a fail-open severity, but an EACCES or corrupted-index failure is then displayed with a false cause, and the advisory warning is silently skipped. Pattern: when a check cannot distinguish failure modes, use a generic "not checked: <tool> failed: <error>" message (surfacing the error value) instead of naming the most common cause; reserve cause-specific messages for causes actually verified (e.g., exit status or stderr markers).
