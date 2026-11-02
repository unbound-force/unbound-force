---
tag: scaffold-drift
author: jay-flowers
category: gotcha
created_at: 2026-11-02T04:22:31Z
identity: scaffold-drift-20261102T042231-jay-flowers
tier: draft
---

When a sibling repo migrates its Go module identity to a major-version suffix (github.com/unbound-force/gaze → github.com/unbound-force/gaze/v2), unsuffixed `go install` commands in UF-owned scaffolded templates do not fail: `@latest` silently resolves to the last pre-migration tag (v1.9.0), so users receive a stale binary with no error signal. Verified while triaging unbound-force/unbound-force#633: the stale path appears not only in the two named `.opencode/` files but also in their byte-identical embedded mirrors under `internal/scaffold/assets/opencode/` (drift-detection pairs — each mirror must be updated in the same change as its canonical `.opencode/` copy or asset-parity tests fail), in `.opencode/agents/gaze-reporter.md` (Gaze-init-scaffolded content with no UF mirror — fix belongs in the Gaze repo), and in executable paths (the Makefile ensure-gaze target; internal/setup/setup.go's installViaGo fallback, which is release-gated: the /v2 path cannot be switched until a corrected stable /v2 tag is published). Module-path migration sweeps should grep the whole repo for the old path and fix each occurrence according to its ownership model.
