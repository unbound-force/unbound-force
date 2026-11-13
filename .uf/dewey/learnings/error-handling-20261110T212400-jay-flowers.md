---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-10T21:24:00Z
identity: error-handling-20261110T212400-jay-flowers
tier: draft
---

Line-anchoring an output-signature classification (`(?m)^fatal: ...`) closes mid-line spoofing from paths that embed the phrase, but not newline-embedded spoofing: a path containing a newline followed by the signature puts the phrase at a line start when the tool echoes the path inside a different error, flipping a real failure into the benign-skip branch. Close the residual by rejecting control characters in the classified input path before invoking the subprocess, or by anchoring the full parenthesized message (`fatal: not a git repository (or any of the parent directories)`), accepting git-version wording risk. Residual observed on checkTrackedPluginLock (unbound-force PR #1006, internal/doctor/checks.go) after its LC_ALL=C + line-anchor hardening.
