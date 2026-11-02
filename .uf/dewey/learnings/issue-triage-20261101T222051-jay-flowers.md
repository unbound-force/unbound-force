---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:20:51Z
identity: issue-triage-20261101T222051-jay-flowers
tier: draft
---

Documentation-gate issues claiming a change needs AGENTS.md coverage are verifiable with three cheap repository checks: (1) the originating change's tasks.md gate task enumerates the same user-facing surfaces the issue lists (derive-per-run-timeout task 8.3 lists the input contract, output fields, RD-FR-009 config fields, and command prose; issue #859 enumerates the same set), (2) the claimed surfaces actually shipped (grep for the config fields hits the three schema copies: schemas/, .uf/schemas/, internal/scaffold/assets/uf/schemas/), and (3) the documentation gap is real (grep AGENTS.md for the surface terms returns zero matches). An issue passing all three is a VALID/enhancement/objective discharge of a mandated gate obligation, not an aspirational ask. Note: the in-repo AGENTS.md documentation gate and the org constitution's Cross-Repo Documentation rule (issue in unbound-force/website) are separate deliverables — one issue cannot discharge both, and filing one does not invalidate the other.
