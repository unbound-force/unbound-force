---
tag: command-authoring
author: jay-flowers
category: gotcha
created_at: 2026-11-05T22:22:47Z
identity: command-authoring-20261105T222247-jay-flowers
tier: draft
---

Double-quoting interpolated refnames in command-prose shell templates is a partial mitigation, not neutralization: inside double quotes bash still expands $, $( ), and backticks, and a literal " ends the quoted span — and $, backtick, and " are all legal git refname characters (git-check-ref-format prohibits only control characters, space, and ~ ^ : ? * [ \). A PR-author-controlled branch name like x$(cmd)y interpolated into `git config "branch.<HEAD_BRANCH>.remote"` executes command substitution when the agent renders the template, and string-equality cross-checks (HEAD_MERGE_REF == refs/heads/<head_ref>) pass the crafted name through. Until an argument-array tool (#965) removes string interpolation entirely, gate interpolation on a conservative charset (e.g. ^[A-Za-z0-9._/-]+$ on HEAD_BRANCH/HEAD_MERGE_REF) — git check-ref-format is not sufficient because the dangerous characters are legal refnames. Found verifying PR #955 iteration-2 quoting correctness in uf.address-feedback Section 1.2.1/4.4.
