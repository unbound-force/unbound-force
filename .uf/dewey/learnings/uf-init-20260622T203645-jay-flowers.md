---
tag: uf-init
author: jay-flowers
category: gotcha
created_at: 2026-06-22T20:36:45Z
identity: uf-init-20260622T203645-jay-flowers
tier: draft
---

The AGENTS.md behavioral rule "Never use git add -A or git add . on feature branches -- stage files explicitly" applies to all branches including opsx/ branches. During spec review of the fix-uf-init-customizations change, the Architect reviewer caught that the commit-before-archive insertion content used git add -A, which directly contradicts this rule. The fix was to use explicit staging: git add openspec/changes/<name>/ .opencode/ plus any other modified files shown by git status. This was a HIGH severity finding because the behavioral rules are described as non-negotiable with CRITICAL severity for violations.
