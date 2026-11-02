---
tag: scaffold-governance
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:09:30Z
identity: scaffold-governance-20261102T040930-jay-flowers
tier: draft
---

In the uf scaffold engine (internal/scaffold/scaffold.go), the file-ownership model hides a force-path exception behind a shared skip message: isNeverOverwrite (currently only specify/memory/constitution.md) is checked before the opts.Force branch, so --force overwrites every asset EXCEPT the constitution — yet the summary prints one "skipped: N (use --force to overwrite)" line for all skip dispositions (scaffold.go printSummary) and the flag help says "Overwrite all existing files" (cmd/unbound-force/main.go). The never-overwrite guard is a governance-integrity control: the constitution is each repo's highest-authority artifact (hero repos carry customized MUST rules with parent_constitution references), and removing the guard would let a routine `uf init --force` silently revert customized governance to the starter template outside the PR+review amendment process — a gate-tampering vector. Routing rule for complaints about this messaging: fix the skip reason and --force help text to name the protection (maintainer-selected Option B on issue #585), never remove isNeverOverwrite. Complements learning scaffold-ownership-20261102T035446-jay-flowers, which covers the user-owned class (issue #649).
