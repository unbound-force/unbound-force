---
tag: feedback-triage
author: jay-flowers
category: gotcha
created_at: 2026-11-06T18:40:52Z
identity: feedback-triage-20261106T184052-jay-flowers
tier: draft
---

Verifying a cited convention threshold by keyword grep can produce a false "unwritten rule" finding. default.md CS-004 phrases its duplication threshold as "Identical or nearly identical logic appearing in more than two locations MUST be extracted into a shared function, method, or module" — greps for "threshold", "three", or "occurrence" miss the phrase "more than two locations", so a triage that grepped instead of reading concluded no written occurrence threshold existed and treated the author's three-or-more citation as fabricated (PR 948 thread PRRT_kwDORVoNbc60JEsN, 2026-11). Rule-citation verification for feedback triage should read the cited rule's full text in the pack file; when a trusted fact-block and the pack text disagree, the pack file on disk is the tiebreaker. The three-signal DRY triage heuristic (feedback-triage-20261029T133813) names signal 1 as "check the convention pack's extraction mandate" — this gotcha pins HOW to check it.
