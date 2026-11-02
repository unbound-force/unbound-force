---
tag: scaffold-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:22:42Z
identity: scaffold-drift-20261102T042242-jay-flowers
tier: draft
---

Cross-repo Go module-path migrations break scaffolded install guidance silently: after a sibling repo migrates its module identity to a major-version suffix (github.com/unbound-force/gaze → github.com/unbound-force/gaze/v2), scaffolded text still recommending the unsuffixed path does not error — `go install <unsuffixed>/cmd/...@latest` resolves to the last v1 tag, installing an outdated major version with no failure signal. Fix at the scaffold owner's source templates together with their embedded mirrors (UF: .opencode/ canonical copies paired byte-identically with internal/scaffold/assets/opencode/ mirrors, parity enforced by drift-detection tests); sibling-owned scaffolded copies (e.g., .opencode/agents/gaze-reporter.md, deployed by `gaze init`) refresh via the sibling's smart-sync rather than local edits. Corrected guidance should mirror the sibling's own README install section — primary package manager, /v2 `go install` only after a verified corrected stable release, prereleases by explicit tag only, release binaries as fallback. Upstream PR "Known issues" sections reliably record these cross-repo fix deferrals (gaze PR #296 → unbound-force/unbound-force#633).
