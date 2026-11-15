---
tag: determinism
author: jay-flowers
category: gotcha
created_at: 2026-11-15T05:06:20Z
identity: determinism-20261115T050620-jay-flowers
tier: draft
---

When a bug report shows an agent executing a prose decision wrongly, fix it at the tool layer, not by writing more precise prose. PR #1077 (fixing issue #1055: /uf.review-pr recommended skip for a stale same-verdict review) rewrote the wrong recommendation as a new prose classification step — commit_id == head_sha comparison plus boolean routing between two gates — even though resolve_review_mode, invoked earlier in the same command, already finds the latest review, compares commits, and returns prior_findings. Before adding procedural prose to a command, run the owner lookup (node .opencode/lib/tool-inventory.ts --json --match "<step words>") for the step's mechanical conjuncts, and check whether a tool the command already calls computes the same inputs. Mixed judgment/mechanical steps split at the DE-006 boundary: mechanical conjuncts (comparisons, boolean routing, gate selection) move into the tool; judgment conjuncts (e.g. "findings materially changed") arrive as validated parameters, and only presentation/recommendation wording stays in prose.
