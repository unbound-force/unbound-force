---
tag: instruction-lint
author: jay-flowers
category: gotcha
created_at: 2026-11-09T19:42:43Z
identity: instruction-lint-20261109T194243-jay-flowers
tier: draft
---

instruction-lint findings attribution requires the scoped run, not the pre-flight's unscoped run. When pre-flight falls back to an unscoped lint (e.g., merge-base fallback against a non-ancestor base such as upstream/main), it reports repo-wide pre-existing lint debt across all 90 instruction files. Attributing findings to a PR requires `.uf/scripts/instruction-lint --json --base main --changed`: its delta.files list and findings are the changed-file record. On PR #1000 (docs-only, six .md files outside .opencode/), the scoped run returned delta.files: [] and findings: [] while the pre-flight's unscoped advisory run reported 1 error + 31 warnings + 24 info — all pre-existing, none PR-caused. Reviewers reading only the pre-flight output would misattribute that debt to the PR.
