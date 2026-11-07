---
tag: test-quality
author: jay-flowers
category: pattern
created_at: 2026-11-07T16:07:35Z
identity: test-quality-20261107T160735-jay-flowers
tier: draft
---

To pin a could-not-run exit code that must differ from the findings exit code (e.g. exit 2 vs exit 1), force a crash inside the REAL entry file rather than mocking the failure: create a dangling symlink where the tool reads files (readFileSync throws ENOENT), spawn the actual entry via execFile, and assert the process exits 2 with a could-not-run message on stderr. This verifies the top-level catch actually separates crashes from findings — a mocked or library-level test cannot, because the separation lives in the entry point's error boundary. Pair it with subprocess tests for the dependency-pre-check wrapper paths (no entry file, no node_modules, partial install) so every lane of the exit contract is exercised end-to-end.
