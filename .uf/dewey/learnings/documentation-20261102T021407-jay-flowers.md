---
tag: documentation
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:14:07Z
identity: documentation-20261102T021407-jay-flowers
tier: draft
---

User-facing docs/ enumerations of scaffolded assets are a second, un-enforced inventory alongside AGENTS.md's authoritative list. In unbound-force/unbound-force, the convention-pack listings in docs/architecture.md (~line 218, "Scaffolded Files" → Convention packs, and ~line 708, "Pack Types") and docs/usage.md (~line 243, "Customization") omit both determinism.md (determinism-codification, PR #723) and meta-prompt.md/meta-prompt-custom.md (model-matrix-scaffold-deployment), while AGENTS.md's Convention Packs section lists all 18 pack files. Each adding change updated the AGENTS.md listing but not the docs/ enumerations, and no drift test binds docs listings to the scaffold manifest, so they drift silently. Recipe: any change that adds a pack, skill, or scaffolded asset MUST grep docs/ for the enumeration sections (architecture.md Scaffolded Files and Pack Types, usage.md Customization, cli-reference.md uf init prose) and update or explicitly scope them within the same change. Prevention: prefer pointer-style references ("see .opencode/uf/packs/") over enumerated file listings in user-facing docs — the same count-free discipline that prevents ordinal-principle drift in instruction files.
