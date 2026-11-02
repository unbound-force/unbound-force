---
tag: go-stdlib-gotchas
author: yvonne-devlin
category: gotcha
created_at: 2026-11-02T17:25:23Z
identity: go-stdlib-gotchas-20261102T172523-yvonne-devlin
tier: draft
---

Go's filepath.Glob ignores filesystem I/O errors such as permission-denied directories and returns empty results indistinguishable from no-match. CLIs that must tell "absent" from "unreadable" (gates, validators, linters) need an explicit os.Stat or os.ReadDir probe per candidate path rather than Glob-only discovery. Observed in uf-gate (PR #582): globFirst (internal/gate/checks.go:101) discards the Glob error channel, so a permission failure during artifact discovery collapses into a "No artifact found" check failure, while direct reads in loadTasksFile surface "Cannot read" messages — one environmental failure class, three different presentations, none matching the spec's exit-2 internal-error contract (FR-009).
