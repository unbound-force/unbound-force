---
tag: seam-verification
author: jay-flowers
category: pattern
created_at: 2026-11-08T18:08:33Z
identity: seam-verification-20261108T180833-jay-flowers
tier: draft
---

Convention-keyed evidence tools — tools that credit test coverage by matching a hardcoded filename and entry shape in a test file they do not own — need a discoverability test that runs the extraction against the real repository and asserts the expected credit. Scratch fixtures written in the same change as the extraction regex verify the regex but cannot detect drift between the convention and the real layout (rename, retirement, or shape change of the credited suite). Verified instance (unbound-force PR #959): tool-inventory.ts seamReport() credits the DE-007 registration layer only to registration-surface.test.ts `{ id: "..." }` entries; a committed learning claimed the set resolves empty for the repo (false — the pre-existing suite matches the regex), and the repo-level discoverability test the learning recommended was not shipped. Without the pin, a suite rename or shape change silently turns every future plugin-tool change into a false `missing: ["registration"]` CRITICAL DE-007 finding.
