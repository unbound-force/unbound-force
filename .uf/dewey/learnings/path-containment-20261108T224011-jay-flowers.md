---
tag: path-containment
author: jay-flowers
category: gotcha
created_at: 2026-11-08T22:40:11Z
identity: path-containment-20261108T224011-jay-flowers
tier: draft
---

Static-analysis tools that follow import specifiers (or any file-content-derived path) must contain their reads within the workspace root. In agent-review flows the analyzed files are untrusted PR content: a plugin or module added by a PR controls the relative specifier, and `join(root, rel)` with a `..` chain reads outside the repo when the reviewer runs the tool on the PR checkout. The fix is a containment check at the reader — canonicalize `resolve(root, rel)` and reject anything that does not stay under root — not at each call site. PR #959 round 3: tool-inventory.ts locateFactory/fsReader shipped without it (Adversary MEDIUM finding).
