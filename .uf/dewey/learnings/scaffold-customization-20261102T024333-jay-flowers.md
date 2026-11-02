---
tag: scaffold-customization
author: jay-flowers
category: gotcha
created_at: 2026-11-02T02:43:33Z
identity: scaffold-customization-20261102T024333-jay-flowers
tier: draft
---

Customizations layered onto upstream-generated files via text-anchor injection break silently on upstream version upgrades: the anchors match the old stock text, so post-init injections no-op with no error signal. Seen in unbound-force issue #713 (openspec 1.2.0 -> 1.14.1): /uf.init injection anchors matched 1.2.0-era stock text and would silently fail against 1.14.1 output. Mitigations: pin scaffold expected-file checks to a stable core set (warning, not failure, for missing expanded-set files such as legacy-migrated profiles lacking the update workflow); treat regeneration that clobbers customizations as expected and re-apply via an explicit KEEP/DROP reconciliation table; longer term, decouple customizations from stock file internals with wrapper files/skills so upstream regeneration cannot break injections.
