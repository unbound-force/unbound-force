---
tag: supply-chain
author: jay-flowers
category: gotcha
created_at: 2026-11-10T20:05:24Z
identity: supply-chain-20261110T200524-jay-flowers
tier: draft
---

Tracker issues filed from review-council findings transcribe verification evidence (git heads, integrity hashes) from the PR under review, creating an ungated second copy of the provenance record. When the work moves to a different PR — unbound-force effect override: #1034 filed against PR #1024, work landing in PR #1026 — the transcribed values can silently contradict the record that ships (#1034 cites git head dff25449… while PR #1026's provenance record carries ec94bd69… for effect@4.0.0-beta.103, both 'verified 2026-11-10'). Curator reviews of security-pin PRs should diff the tracker's cited evidence against the PR's actual provenance record before merge: a mismatch means the Gatekeeping 'verified upstream SHA' condition has contradictory evidence until one artifact is corrected.
