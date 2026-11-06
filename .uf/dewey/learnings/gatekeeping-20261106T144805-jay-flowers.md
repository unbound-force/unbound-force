---
tag: gatekeeping
author: jay-flowers
category: pattern
created_at: 2026-11-06T14:48:05Z
identity: gatekeeping-20261106T144805-jay-flowers
tier: draft
---

When a pinned-SHA fix documents its rollback by retaining the previous SHA in the CHANGELOG and spec artifacts, the accompanying verification task must scope the old-SHA-absence grep to functional locations (.github/workflows), not repo-wide: a repo-wide grep always fails against the intentional rollback record. Phrase task criteria to match the check actually executed — a checkbox whose written criterion is unsatisfiable against the same PR's rollback documentation is a false record even when the properly scoped check passed. Observed in fix-pinact-gaze-pin tasks.md 3.1 (marked [x] while its repo-wide criterion contradicts tasks 3.3/4.1 and the CHANGELOG entry in the same PR).
