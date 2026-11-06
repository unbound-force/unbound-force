---
tag: determinism
author: jay-flowers
category: pattern
created_at: 2026-11-06T21:12:31Z
identity: determinism-20261106T211231-jay-flowers
tier: draft
---

DE-004 change records should paste the instruction linter's full `delta.files` output, not just the headline files: every changed instruction file needs its line/token delta row, including small edits such as AGENTS.md convention updates or +2-line persona tweaks, and the rows should be re-run after final-round edits so recorded counts match head. The linter's `--json` delta block (totals plus per-file lines/tokens/loadTokens) is the paste-ready format; partial or stale size records leave the efficiency claim only partially evidenced.
