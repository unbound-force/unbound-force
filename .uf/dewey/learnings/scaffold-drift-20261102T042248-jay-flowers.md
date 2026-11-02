---
tag: scaffold-drift
author: jay-flowers
category: pattern
created_at: 2026-11-02T04:22:48Z
identity: scaffold-drift-20261102T042248-jay-flowers
tier: draft
---

When a sibling repository migrates its Go module to a major-version suffix path (github.com/unbound-force/gaze → github.com/unbound-force/gaze/v2), every unsuffixed `go install` reference in the UF meta repo silently goes stale: unsuffixed `@latest` freezes at the last v1 tag, so users following the guidance install the outdated major version. The stale path typically appears in more surfaces than a report names — canonical `.opencode/` templates, their byte-parity mirrors under `internal/scaffold/assets/opencode/` (drift-detection tests fail if only one side of a pair is fixed), UF-internal install paths (Makefile `ensure-gaze` target, `internal/setup/setup.go` go-install fallback), and sibling-owned copies refreshed only by that sibling's init command. Triage and fix sweeps for module-path migrations should grep the module path repo-wide (both the unsuffixed and suffixed forms) rather than trust the named-file list in the report.
