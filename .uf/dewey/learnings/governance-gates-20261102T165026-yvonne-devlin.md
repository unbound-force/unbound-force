---
tag: governance-gates
author: yvonne-devlin
category: gotcha
created_at: 2026-11-02T16:50:26Z
identity: governance-gates-20261102T165026-yvonne-devlin
tier: draft
---

When tightening artifact resolution in a multi-change repository from first-match globs to strict branch-aware lookup, content-scanning checks are the commonly missed outlier: path-existence checks receive the strict resolver while keyword/coverage scanners keep unscoped `base/*/artifact` globs and remain satisfiable by any sibling change — a silent false pass in a governance gate. Whenever resolution semantics change, audit every check that reads repository artifacts (not only those whose return value is a path) and add a branch-mapped multi-change regression test per check, mirroring the resolver tests.
