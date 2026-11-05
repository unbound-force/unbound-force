---
tag: spec-governance
author: jay-flowers
category: pattern
created_at: 2026-11-05T18:34:55Z
identity: spec-governance-20261105T183455-jay-flowers
tier: draft
---

A MODIFIED requirement's Previously paragraph that quotes values an intermediate unarchived change moved (e.g., a rollback SHA updated from b628c764... to 5ebee4af... between the root requirement's author and the current change) creates a hidden chain dependency: the delta silently anchors to the intermediate change's text, not the named predecessor's. Verified on opsx/stop-shipping-plugin-lock: its SC-FR-003 Previously paragraph quotes clarification-marker-superset's moved SHA while the delta header names review-council-multi-model-fanout as the amended change.
