---
tag: instruction-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-16T01:48:55Z
identity: instruction-quality-20261116T014855-jay-flowers
tier: draft
---

Partially typed tool-input bullet lists create invalid_input traps. When most parameters in a command's input-contract bullets carry parentheticals (nullable, ≥ 1, 7–40 hex, enums) and a few stay bare, executing agents map from earlier singular definitions in the same command (e.g., an assessment table's single "Evidence" reference) into array-typed tool fields and fail validation. Type every field in the bullet list, or point at the schema instead of typing only some.
