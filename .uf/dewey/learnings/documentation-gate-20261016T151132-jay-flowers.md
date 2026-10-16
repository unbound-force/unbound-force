---
tag: documentation-gate
author: jay-flowers
category: gotcha
created_at: 2026-10-16T15:11:32Z
identity: documentation-gate-20261016T151132-jay-flowers
tier: draft
---

The unbound-force/unbound-force repository has coexisting `docs` (12 open issues) and `documentation` (8 open, including #717, #720, #736, #740) labels, so single-label queries miss half the documentation-gate namespace. Reviewers verifying doc-gate compliance must query both labels or resolve cited issue numbers directly via `gh issue view N`, using the CHANGELOG `Docs: #N` trailer and tasks.md check-off notes as anchors. Issue filers should add the `docs` label at creation time; this is the same failure class as unbound-force/website#303.
