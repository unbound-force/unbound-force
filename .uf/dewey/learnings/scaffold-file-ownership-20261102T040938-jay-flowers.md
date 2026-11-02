---
tag: scaffold-file-ownership
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:09:38Z
identity: scaffold-file-ownership-20261102T040938-jay-flowers
tier: draft
---

Skip/disposition output in the uf scaffold engine must be protection-class-aware. The engine has three effective file classes — tool-owned (auto-update on content diff), user-owned (skip; --force overwrites), and never-overwrite (skip; --force does NOT overwrite) — but the skip summary emits one generic "(use --force to overwrite)" hint for both skip classes. Because isNeverOverwrite (currently only .specify/memory/constitution.md, evaluated before the opts.Force branch in internal/scaffold/scaffold.go) makes that hint false for the protected class, users following it hit a dead end (unbound-force issue #585). Maintainer direction on record: the protection is by design; fix the message (Option B — distinguish "temporarily skipped, use --force" from "protected governance file, not overwritten"). Implementer corollary: the completed-but-unarchived openspec/changes/fix-constitution-scaffold delta (fixes #213) still carries an FR-003 "Force mode overwrites constitution" scenario that predates the isNeverOverwrite guard and contradicts both the implemented behavior and the maintainer's Option B choice — reconcile that delta when fixing the message, or the spec of record will codify force-overwrite semantics the maintainer explicitly declined. Related but distinct from the #649 user-owned-class finding (review-matrix.yaml): same generic-hint mechanism, different protection class.
