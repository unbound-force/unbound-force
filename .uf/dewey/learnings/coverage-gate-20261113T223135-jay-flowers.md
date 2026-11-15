---
tag: coverage-gate
author: jay-flowers
category: gotcha
created_at: 2026-11-13T22:31:35Z
identity: coverage-gate-20261113T223135-jay-flowers
tier: draft
as_of: "unbound-force@4558c49ae60be6c02a7f2d23c541518ee50d7fd4"
---

A changed-package coverage gate reads `git diff` paths, but the Go tool never measures sources under dot-prefixed or underscore-prefixed directories or `testdata` (per `go help packages`). Any fixture or generated tree that lives under a dot directory (e.g. `.eval/lib/mini-repo/**/*.go`) therefore appears as a changed package that can never appear in a `go test ./...` coverprofile, so every PR touching it fails the coverage gate as unmeasurable. The fix shipped in PR #1061 is to mirror the go tool's wildcard-ignore rules in `PackagesFromFiles` (leading `.`/`_` segments, `testdata` segment), keeping the gate's package set aligned with what the module's own test run can measure. Check both directions when adding such trees: the gate must not demand coverage it cannot obtain, and the filter must not hide buildable production packages.
