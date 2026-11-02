---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-02T03:33:35Z
identity: spec-authoring-20261102T033335-jay-flowers
tier: draft
---

The host-fallback-provenance OpenSpec proposal demonstrates the documentation-complete spec-phase pattern for unbound-force changes: an Impact-section \"Documentation impact\" block that names every documentation surface — CHANGELOG.md entry, the targeted docs/ section (naming the existing section heading), the in-repo documentation issue per the AGENTS.md documentation-gate precedent, and the unbound-force/website issue per the org constitution's Cross-Repo Documentation rule — cites precedent issue pairs (#778 ↔ website#316, #891 ↔ website#360), and tasks each item in tasks.md with PR-body linkage. Reviewing specs against this checklist is mechanical: every user-facing behavior change should map to a tasked CHANGELOG entry, a tasked docs note or an explicit no-change assessment for AGENTS.md/README.md, and tasked in-repo + website documentation issues.
