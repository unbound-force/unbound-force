---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-06T18:48:57Z
identity: spec-governance-20261106T184857-jay-flowers
tier: draft
---

When an OpenSpec delta hardcodes dependency version literals, a pin bump executed through the documented gatekeeping security path must sweep every literal site in the unarchived change before archive: the requirement's pin sentence, override-enumeration sentences ("The toml and uuid security overrides MUST remain"), scenario examples that embed the pin as fixture data, and the proposal's summary bullets. Tests catch the manifest and provenance sides (pin cross-checks, override-map assertions) but no test catches spec-text literals — a partial amendment silently archives a self-contradicting spec. Seen in the stop-shipping-plugin-lock zod 4.1.8→4.6.6 bump plan: the delta carries the pin at three sites, the plan named one.
