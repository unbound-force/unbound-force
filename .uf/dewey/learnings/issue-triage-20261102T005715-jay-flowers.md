---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T00:57:15Z
identity: issue-triage-20261102T005715-jay-flowers
tier: draft
---

Phased migration umbrellas require duplicate checks that distinguish phase siblings from true duplicates. A phase issue (e.g., "Phase 1 — CI v2 leg" under an OpenCode v2 migration umbrella) is not a duplicate of its umbrella, its predecessor spike, or a complementary issue even when all carry the same label and topic. Classifying signals: the phase issue names its umbrella and phase position, states the sibling relationship and deliverable boundary ("complements #714, whose validation plan is manual" — continuous CI verification versus plugin code changes), and carries phase-scoped, testable "Done when" criteria. An issue that overlaps a sibling's deliverable while lacking these boundary signals is the actual duplicate candidate.
