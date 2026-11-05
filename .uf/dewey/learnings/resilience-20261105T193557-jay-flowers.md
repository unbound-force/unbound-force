---
tag: resilience
author: jay-flowers
category: gotcha
created_at: 2026-11-05T19:35:57Z
identity: resilience-20261105T193557-jay-flowers
tier: draft
---

Recovery-path gates must accept the state the recovery path itself produces. /uf.address-feedback's new Section 1.2.1 gated re-entry on exact HEAD == frozen head_sha equality, while the command's own Section 4.4 push-failure path preserves unpushed local commits — leaving the PR head branch strictly ahead of the frozen SHA. After any commit-then-failed-push run, every rerun STOPped at the checkout gate before get_progress could resume, making the checkpoint/resume feature unreachable for its primary failure mode. Rule: when a workflow preserves local-ahead state for retry, gate re-entry on containment (`git merge-base --is-ancestor <remote_sha> HEAD`), not equality; reserve exact-equality gates for cold starts that carry no preserved state.
