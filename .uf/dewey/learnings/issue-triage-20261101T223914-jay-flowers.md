---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:39:14Z
identity: issue-triage-20261101T223914-jay-flowers
tier: draft
---

Documentation-impact issues in unbound-force/unbound-force carry count-bump dispositions that function as verifiable drift anchors: the AGENTS.md Project Structure slash-commands count (e.g. 32→33 files), the QUICKSTART.md scaffolded-commands count and the expectedAssetPaths embedded-asset registration (e.g. 11→12 each), and docs/architecture.md's Scaffolded Files total (e.g. ~46→~47). Triage and review should verify these three ways: the counts are mutually consistent within the issue, they match the landed AGENTS.md/QUICKSTART.md/docs text after the PR, and they equal the number of files the change actually added. A mismatch in either direction — the issue claims a bump the repository does not show, or the repository shows a bump the issue does not claim — is a documentation-drift signal, not a formatting nit. Observed passing case: issue #862 (uf-fix-ci docs impact), whose 32→33, 11→12, and ~46→~47 bumps all match the post-PR AGENTS.md state and the change's tasks.md documentation phase (tasks 3.1–3.3).
