---
tag: error-handling
author: jay-flowers
category: gotcha
created_at: 2026-11-07T19:03:56Z
identity: error-handling-20261107T190356-jay-flowers
tier: draft
---

Classifying subprocess failures by matching a diagnostic signature in combined output must anchor to line position, not substring presence: git and similar tools interpolate the caller-controlled target path into other fatal messages ("fatal: cannot change to '<path>'"), so even a full-phrase Contains match like "fatal: not a git repository" is spoofable by a target path that embeds the phrase, converting a real failure into the benign skip. Precedent: checkTrackedPluginLock (internal/doctor/checks.go, PR #957 feedback fixes) — anchoring to the "fatal:" prefix narrowed the masquerade its comment claims to close but did not eliminate it; the robust form accepts only a line whose trimmed text starts with the signature (or classifies on exit code plus first stderr line). Audit rule: whenever failure classification reads combined output that echoes an input-influenced path, prefer per-line prefix matching over Contains.
