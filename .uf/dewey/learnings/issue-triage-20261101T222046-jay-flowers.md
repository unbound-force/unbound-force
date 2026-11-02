---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:20:46Z
identity: issue-triage-20261101T222046-jay-flowers
tier: draft
---

Documentation-gate issues filed per the AGENTS.md gate triage fast when they carry three verifiable elements: the originating change directory (e.g., openspec/changes/&lt;name&gt;/), an enumeration of each user-facing surface with its governing requirement ID (e.g., RD-FR-009, DR-FR-002), and the CHANGELOG entry location. With those present, triage can confirm gate applicability (user-facing change outside the exempt set: internal refactoring, test-only, CI-only, spec artifacts) and proportionality (one change, one doc deliverable) without reopening the change artifacts. The proportionate outcome is VALID / enhancement / objective with no split — splitting one change's gate-tracking issue fragments the gate it exists to satisfy.
