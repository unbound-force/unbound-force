---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:54:07Z
identity: issue-triage-20261101T225407-jay-flowers
tier: draft
---

Documentation-impact issues filed as tasks inside a spec change (e.g., unbound-force/unbound-force#844, task 3.2 of the replicator-version-floor OpenSpec change) operationalize the AGENTS.md documentation gate and triage as VALID / enhancement / objective when they contain per-surface dispositions (CHANGELOG.md, docs/* pages, AGENTS.md, README.md), explicit blog/tutorial significance decisions, and acceptance checkboxes tied to the implementing PR — with every factual claim corroborated by the change artifacts under openspec/changes/<name>/ (Dewey indexes them; verify the body against proposal.md, design.md, and the specs/ delta before ruling). The recurring open item is the website disposition: the org constitution's Cross-Repo Documentation rule requires a unbound-force/website issue for user-facing CLI output changes before PR merge, so an in-repo docs issue that leaves it conditional ("link one here if the website documents X") is tracking a real obligation the implementing PR must resolve — treat it as a confirmation item, not a triage blocker.
