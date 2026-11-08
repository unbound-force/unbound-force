---
tag: ci-cd
author: jay-flowers
category: pattern
created_at: 2026-11-05T19:53:39Z
identity: ci-cd-20261105T195339-jay-flowers
tier: draft
---

When deleting a tracked lockfile, audit every consumer of the artifact for detection-cadence parity, not just content parity. A daily scheduled scanner plus Dependabot provide time-based CVE detection for a locked tree; a compensating throwaway-lock scan added only to a push/PR-triggered workflow does not restore that cadence — a vulnerability disclosed after the last merge goes undetected until the next push. Compensating controls must inherit the trigger cadence of the control they replace (e.g., wire the compensating job into the scheduled workflow too).
