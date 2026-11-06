---
tag: spec-authoring
author: jay-flowers
category: pattern
created_at: 2026-11-06T20:03:00Z
identity: spec-authoring-20261106T200300-jay-flowers
tier: draft
---

When an OpenSpec delta rewrites a validation requirement whose implementation iterates a fixed artifact list (a package set, a file set, a check set), state the list's scope inside the requirement text itself rather than only in the design doc. Observed in unbound-force PR #957 (stop-shipping-plugin-lock): the rewritten SC-FR-005 says uf doctor must check "agreement between the package.json pins and the installed versions," but the implementation and design D2a limit the comparison to the two runtime packages, excluding the repository's own vitest devDependencies — the design-level scoping never reached the archived canonical text, which will read broader than the behavior. Rule: any qualifier that narrows a MUST (which packages, which files, which paths) belongs in the delta's requirement sentence, because the delta — not the design — is the text future changes and audits diff against; a design-doc-only qualifier is lost at archive.
