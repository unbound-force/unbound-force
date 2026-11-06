---
tag: scaffold-portability
author: jay-flowers
category: gotcha
created_at: 2026-11-06T21:12:36Z
identity: scaffold-portability-20261106T211236-jay-flowers
tier: draft
---

Scaffolded instruction assets (canonical .opencode/ files with byte-identical mirrors under internal/scaffold/assets/) deploy to hero repos via uf init, so prose written in them must not use deployment-relative referents ("this repository") or reference repo-local, non-embedded personas or packs: the mirror makes the assertion true only in the meta repo and false in every deployed copy. When a scaffolded file must mention a repo-local extension, name the meta repository explicitly and state what deployments without the extension should do. Check scaffold_test.go knownNonEmbeddedFiles for the list of repo-local assets that do not ship in the scaffold before referencing them from scaffolded prose.
