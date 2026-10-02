---
tag: review-pr-marker-fix
author: jay-flowers
category: gotcha
created_at: 2026-09-29T20:38:56Z
identity: review-pr-marker-fix-20260929T203856-jay-flowers
tier: draft
---

When modifying files under .opencode/commands/ that are also embedded in internal/scaffold/assets/, you MUST sync the embedded asset copy to avoid TestEmbeddedAssets_MatchSource drift detection failures. The pattern is: cp .opencode/commands/FILENAME internal/scaffold/assets/opencode/commands/FILENAME. This was needed when adding the workflow-gate marker exception to uf.review-pr.md for issue #539. The drift detection test runs as part of the normal Go test suite and will fail the build if the embedded copy diverges from the source.
