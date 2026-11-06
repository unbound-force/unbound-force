---
tag: test-quality
author: jay-flowers
category: gotcha
created_at: 2026-11-05T17:01:40Z
identity: test-quality-20261105T170140-jay-flowers
tier: draft
---

An all-failures negative test for an idle/no-progress watchdog proves that failed probes do not ADVANCE the idle clock, but not that they do not RESET it — when every probe fails, 'reset lastProgressAt on failure' and 'ignore failure' are observationally equivalent (no abort either way). To lock the 'failed probe must not reset the clock' semantics, use a mixed sequence: successful probes showing progress, then several failing probes, then a successful unchanged probe — and assert the idle abort measures fr
