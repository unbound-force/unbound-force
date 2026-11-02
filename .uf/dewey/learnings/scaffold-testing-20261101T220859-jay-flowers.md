---
tag: scaffold-testing
author: jay-flowers
category: gotcha
created_at: 2026-11-01T22:08:59Z
identity: scaffold-testing-20261101T220859-jay-flowers
tier: draft
---

Embed-FS test seams require slash-separated path construction, not filepath.Join: io/fs paths are always forward-slash and unrooted, so filepath.Join produces OS-specific separators (backslashes on Windows) that fail embed.FS lookups. When a test builds asset paths for an embed.FS reader (e.g. a scaffold assetContent seam), string concatenation with '/' is the correct idiom — production seams like assetContent itself use "assets/" + relPath. Reviewer note: do not flag raw path concatenation against the Go pack's filepath.Join rule (SC-003) without first checking whether the path targets an embed.FS (in-memory, slash-only, concatenation correct) or the real filesystem (filepath.Join correct). Verified during the #900 review-council: TestSCFR009_DeploymentCompletenessGuard's "opencode/uf/packs/" + key + ".md" concatenation matches the production assetContent seam and is Windows-safe precisely because it avoids filepath.Join.
