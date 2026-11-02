---
tag: issue-triage
author: jay-flowers
category: pattern
created_at: 2026-11-02T02:13:46Z
identity: issue-triage-20261102T021346-jay-flowers
tier: draft
---

Documentation-gate record issues (title "docs: record ...", documentation label) filed after a user-facing change are tracking issues for remaining doc work, not duplicates of the implementation issue they reference. Verify them by triangulation: (1) the referenced OpenSpec change directory exists with completed tasks; (2) the CHANGELOG entry exists and typically cross-references the whole gate chain in its trailer (e.g. "Refs: #732, Docs: #740, unbound-force/website#310"), which both confirms the entry and proves the docs issue number was threaded through; (3) the cross-repo website sync issue is discoverable via Dewey's GitHub index (github-<org> source pages) without leaving the detected repository; (4) a repo-wide grep for the behavioral term (e.g. "model-invocable") confirms the doc gap is still outstanding — the pattern from learning/issue-triage-20261101T233554-jay-flowers. When the chain holds and a keyword search returns only the issue itself, triage as VALID / enhancement / objective with null split: the implementation issue (closed) and the gate record (open) are distinct mandated artifacts, and related-topic hits from other changes (e.g. issues #873, #834) are not duplicates. Verified on unbound-force issue #740 (fix-tool-args-schema-mismatch).
