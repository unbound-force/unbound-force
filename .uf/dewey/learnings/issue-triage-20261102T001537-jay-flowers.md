---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:15:37Z
identity: issue-triage-20261102T001537-jay-flowers
tier: draft
---

Documentation-drift pattern in command input contracts (unbound-force/unbound-force chain #763 → #811 → #814): /uf.address-feedback §2.3.4 documented a diff_path input before any step produced it, which required a follow-up change (#811, adding Phase 1 Section 1.4 to call fetch_pr_review_context) and then a documentation issue (#814) to reconcile. Triage signal: documentation issues that cite their originating change, enumerate the concrete facts to record (producer tool, renumbered sections, STOP conditions, tool-enforced size limits), and carry a documentation label are objectively verifiable — accept them when duplicate search is clean and the named command exists in the repo. Prevention signal for command authors: every documented consumer input needs a producing step in the same revision, and section renumbering always breaks cross-references that user-facing docs must track.
