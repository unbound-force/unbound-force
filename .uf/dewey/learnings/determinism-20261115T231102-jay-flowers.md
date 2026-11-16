---
tag: determinism
author: jay-flowers
category: reference
created_at: 2026-11-15T23:11:02Z
identity: determinism-20261115T231102-jay-flowers
tier: draft
---

DE-007's mandatory seam-verification task (registration surface, documented output shape, round-trips) binds prose→tool migrations — increments that extract instruction text into tools and thereby create or change a registration seam. A tool-behavior refinement that changes no plugin registration verifies its changed seams instead — output/sample shape, write→read round-trips, and the documented envelope — through its normal test tasks. Exemplar: compaction-storm-detector-sums-compactio (unbound-force, 2026-11) tasks 3.1–3.3 verify the ProbeCounters sample shape, the TimeoutDetails.storm_session_id envelope, and the round-trip through uf_delegate_command, while the design's Contracts section names exactly those as "the seams verified in tasks.md"; demanding the unchanged registration layer be named would be ceremony.
