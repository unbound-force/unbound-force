---
tag: scaffold-ownership
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:09:40Z
identity: scaffold-ownership-20261102T040940-jay-flowers
tier: draft
---

Scaffold summary output must distinguish skip-reason classes. A generic "(use --force to overwrite)" hint attached to a never-overwrite-protected file (isNeverOverwrite in internal/scaffold/scaffold.go — currently only .specify/memory/constitution.md, checked before the force branch so --force can never reach it) directs users to a flag that cannot succeed, and the same inaccurate contract propagates to the --force flag help text ("Overwrite all existing files") and the website file-ownership docs. Pattern: emit a separate disposition class for permanently protected files (e.g. "protected: governance file, never overwritten") and align every surface that describes --force with its actual reach (everything except isNeverOverwrite entries). Verified during unbound-force issue #585 triage; the maintainer confirmed the guard is intentional and selected the message-fix option, while an in-flight OpenSpec change (fix-constitution-scaffold, FR-003 "Force mode overwrites constitution") still encodes the opposite semantics and needs reconciliation before the fix lands.
