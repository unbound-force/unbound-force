---
tag: scaffold-ownership
author: jay-flowers
category: decision
created_at: 2026-11-02T04:09:35Z
identity: scaffold-ownership-20261102T040935-jay-flowers
tier: draft
---

unbound-force #585 (`uf init --force` skips `.specify/memory/constitution.md`): maintainer decision 2026-10-01 — the `isNeverOverwrite` protection for the constitution is intentional design, and the accepted remedy is the issue's Option B: keep the protection and reword the skip message so permanently protected governance files are distinguished from force-overwritable user-owned files (e.g., "protected (governance file, not overwritten)") instead of the shared "use --force to overwrite" hint; Options A (remove protection) and C (dedicated reset flag) were declined. The inaccurate "--force overwrites all files" claim ships on multiple surfaces the fix should sweep together: the CLI skip message, the `--force` flag help text ("Overwrite all existing files"), and the docs/architecture ownership-model note ("Use `--force` to overwrite all files regardless of ownership"); the website developer.md File Ownership Model documents only the two-class model (tool-owned/user-owned) and omits the never-overwrite constitution class entirely. Reconciliation caveat for implementers: `openspec/changes/fix-constitution-scaffold/specs/scaffold-constitution` FR-003 specifies the opposite semantics ("Constitution Is User-Owned"; force mode overwrites the constitution — Option A), while shipped code still evaluates `isNeverOverwrite` before the force branch (verified during #649 triage, 2026-11); the maintainer's in-issue comment is the direct decision record for this issue's remedy, and the change artifact must be reconciled (superseded or amended) before the message fix lands so spec and code do not diverge.
