---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:13:40Z
identity: issue-triage-20261102T021340-jay-flowers
tier: draft
---

Triage classification for documentation-gate record issues: when an issue is a documentation-gate record (label "documentation"; body names a change, its user-facing impact, and documentation to update), classify VALID / enhancement / objective even though the underlying code fix is already merged — the issue is the in-repo half of a two-half gate (repo AGENTS.md documentation gate plus the org constitution's Cross-Repo Documentation rule), and its actionable core is the remaining user-facing doc-sync work. Verify mechanically with three checks: (1) the referenced change artifacts exist (openspec/changes/<name>/ or specs/<NNN>-*/), (2) the CHANGELOG entry exists under the change name, and (3) the paired unbound-force/website issue exists (gh issue list --repo unbound-force/website --state all --search "<change keyword>"). A missing website half is a gate gap to flag in the triage comment, not grounds for INVALID. Established precedent pairs: in-repo #740 + website#310 (fix-tool-args-schema-mismatch), in-repo #710 + website#306 (fix-tui-git-output-leak), in-repo #791 + website#317 (delegation classification).
