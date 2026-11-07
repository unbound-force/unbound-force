---
tag: learning-capture
author: jay-flowers
category: gotcha
created_at: 2026-11-07T17:09:57Z
identity: learning-capture-20261107T170957-jay-flowers
tier: draft
---

Learning-capture writes can silently truncate artifact bodies mid-token, and the clipped text propagates verbatim into the Dewey index — file and index agree, so drift checks cannot see the defect; only a terminal-punctuation sweep on the last body line surfaces it. Before committing captured learnings, run the completeness loop (for f in .uf/dewey/learnings/*.md; do tail -1 "$f" | grep -qE '[.!?)"]$' || echo "$f"; done) and treat every hit as a restore-from-source task: the full-length text lives only where the proposal was authored (dispatch artifact payload.proposals or the authoring session), because git holds only the clipped state and the index mirrors the clip. Truncation clusters by capture window — one bad window clips a whole batch — so sweep the entire corpus, and after any restoration reload the Dewey index so retrieval serves the restored text.
