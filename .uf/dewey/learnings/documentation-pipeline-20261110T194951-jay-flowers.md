---
tag: documentation-pipeline
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:49:51Z
identity: documentation-pipeline-20261110T194951-jay-flowers
tier: draft
---

When a PR carries multiple change units (the author's original work plus fix-ci repair additions), the documentation gate applies per unit, and the visible signal is sibling CHANGELOG entries with asymmetric docs references: PR #1024's effect-osv-override entry carries "(Docs: #1034)" while the adjacent triage-issue-heredoc entry carries none. Verify each unit's gate cell independently — in-repo docs issue AND website-repo issue per the constitution's Cross-Repo Documentation rule — because applying the gate to the unit that triggered the repair loop while skipping the author's original unit leaves a HIGH gap that persists across council iterations (run 2 flagged it; no issue was filed before head 705244a4).
