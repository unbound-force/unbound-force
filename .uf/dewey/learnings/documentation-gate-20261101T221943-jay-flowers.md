---
tag: documentation-gate
author: jay-flowers
category: pattern
created_at: 2026-11-01T22:19:43Z
identity: documentation-gate-20261101T221943-jay-flowers
tier: draft
---

Documentation-gate triage pattern (verified while triaging unbound-force/unbound-force#873): a governance documentation gate and content-pipeline significance thresholds answer different questions and must be evaluated separately. New plugin tools shipped to every consumer via the scaffold mirror are user-facing capability surface, so the repo documentation gate fires (file the tracking issue with dispositions and a docs-surface assessment) even when the change is an internal extraction that preserves all observable command behavior and the blog/tutorial significance thresholds legitimately do not trip. For cross-repo website tracking, follow the spec-planned search-first rule: run a recorded pass over open unbound-force/website issues, then create a new tracker only when no on-point one exists (fix-review-dispatch-recovery created website#313 when none existed); when on-point trackers exist, reference and consolidate on them (extend-fetch-pr-review-context consolidated on website#342 and #310) rather than duplicating — linked open trackers satisfy the tracking intent of the constitution's cross-repo documentation rule while duplicate filing violates zero-waste. Record each evaluation (gate fired, blog declined, tutorial declined, consolidation) with its threshold reasoning so every disposition is independently auditable.
