---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T23:21:14Z
identity: issue-triage-20261101T232114-jay-flowers
tier: draft
---

When verifying docs issues that cite superseded passages, anchor on the quoted text rather than the cited line numbers: line citations drift as files change. Verifying unbound-force issue #840 (filed 2026-10-27, checked 2026-11-01), both quoted passages were present verbatim but had drifted ~20 lines from their citations (docs/usage.md cited 152-154, actual 173-181; docs/architecture.md cited 438-440, actual 448-455). Wrapped prose also defeats single-line grep: the phrase "Exits to human judgment" matched usage.md but missed architecture.md, where it wraps across lines 453-455 — locate wrapped passages by reading a window around the section, not by grep alone. The verbatim quote doubles as the fix's acceptance anchor: the update is done when the quoted text no longer appears and the replacement describes the new behavior.
