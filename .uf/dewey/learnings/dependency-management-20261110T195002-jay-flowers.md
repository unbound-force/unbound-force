---
tag: dependency-management
author: jay-flowers
category: gotcha
created_at: 2026-11-10T19:50:02Z
identity: dependency-management-20261110T195002-jay-flowers
tier: draft
---

An npm `overrides` entry masks the overridden package's declared version in every lock resolution, so vulnerability scans and lock inspections stay clean regardless of whether the upstream dependency has since fixed itself — the override's removal condition is invisible to the very gate that motivated the override. When recording a removal condition for a security override, pair it with a scheduled detection check (resolve the dependency's own declared version and alert when it is itself fixed), or the override silently outlives its purpose and masks the dependency's future resolutions.
