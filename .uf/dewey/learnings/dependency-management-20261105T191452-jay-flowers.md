---
tag: dependency-management
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:14:52Z
identity: dependency-management-20261105T191452-jay-flowers
tier: draft
---

A spec delta that re-points verification from a live artifact to a static record (provenance JSON) needs a test cross-linking the record to the live source: read the live manifest and compare versions/licenses, or the record goes silently stale on the next bump. A test that only reads the record and compares against hardcoded constants verifies nothing about the record's accuracy.
