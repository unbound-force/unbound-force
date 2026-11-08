---
tag: learning-capture
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:09:59Z
identity: learning-capture-20261107T170959-jay-flowers
tier: draft
---

Dewey learning artifacts captured in batch during review rounds can be truncated at write time, losing each learning's final clause. The clipping is batch-correlated — whole capture batches affected while later batches are complete — and typically cuts mid-token (observed on unbound-force PR #948: 'detached HE' for 'detached HEAD', 'without --ignore-' for '--ignore-scripts'), so a terminal-punctuation check on each artifact's last line (tail -1 "$f" | grep -qE '[.!?)"]$') detects it corpus-wide. Full text remains recoverable from the originating dispatch artifacts' payload.proposals. Remediation must restore the .md artifacts AND refresh the Dewey index (learnings are indexed as graph pages), verifying via retrieval that restored clauses are searchable — a file-only fix leaves semantic retrieval serving clipped content.
