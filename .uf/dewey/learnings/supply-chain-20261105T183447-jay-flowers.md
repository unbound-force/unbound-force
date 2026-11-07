---
tag: supply-chain
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:34:47Z
identity: supply-chain-20261105T183447-jay-flowers
tier: draft
---

When a lockfile removal names package.json `overrides` as a transitive-drift mitigation, the post-install verification surface that replaces lock validation should cover the overridden packages too. npm's override enforcement becomes the only mechanism constraining those transitive versions, and a doctor-style pin-vs-installed check that iterates only direct dependencies leaves the security overrides unverified after install. Seen in stop-shipping-plugin-lock: manifest-install compares the four direct pins but never checks that toml/uuid resolved to the override versions.
