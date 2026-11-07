---
tag: dewey-learnings
author: jay-flowers
category: gotcha
created_at: 2026-11-06T22:14:54Z
identity: dewey-learnings-20261106T221454-jay-flowers
tier: draft
---

Lineage and reconciliation notes must scope absence claims to the exact symbol that was verified, never extend them to the containing file or path. Observed on PR #979 iteration 3 (chore/dewey-learnings-coffee-20261106): the Lineage note in dependency-management-20261106T180908 correctly grep-verified that the symbol preservePluginDevDependencies resolves nowhere in this repo, then over-generalized to "internal/scaffold/plugin_activation.go ... exist[s] only in the coffee workspace's lineage ... neither the function nor the file exists here" — but that file exists in this repo (internal/scaffold/ holds plugin_activation.go and plugin_activation_test.go), is the activation path that runs npm ci, and is cited as local by the compiled corpus (plugin-architecture.md:98), four October learnings, and three unarchived OpenSpec changes. A file named by a foreign-workspace change usually exists locally too — the foreign change modified it, it did not create it. Procedure: after grepping for the symbol, separately verify the cited file path itself (Read on the parent directory, or grep the path string across tracked docs) before writing any file-level absence claim; when only the symbol was verified, write "no source in this repo defines <symbol>" (the scaffold-20261106T180900/180912 pattern), not "the file does not exist".
