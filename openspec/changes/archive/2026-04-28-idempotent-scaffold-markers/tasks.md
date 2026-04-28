## 1. Idempotent marker insertion

- [x] 1.1 Add `stripExistingMarkers(s string) string` to
  `internal/scaffold/scaffold.go`. The function removes all
  lines whose trimmed content starts with
  `<!-- scaffolded by uf ` or `# scaffolded by uf `.
  Preserves all other lines including blank lines.
- [x] 1.2 Modify `insertMarkerAfterFrontmatter` to call
  `stripExistingMarkers` on its input string before
  proceeding with the existing frontmatter detection and
  marker insertion logic.

## 2. Fix double-v version prefix

- [x] 2.1 In `.goreleaser.yaml`, change the `unbound-force`
  build ldflags from `-X main.version={{.Tag}}` to
  `-X main.version={{.Version}}` (line 18).
- [x] 2.2 In `.goreleaser.yaml`, change the `mxf` build
  ldflags from `-X main.version={{.Tag}}` to
  `-X main.version={{.Version}}` (line 35).

## 3. Clean up embedded assets

Strip duplicate markers from all files under
`internal/scaffold/assets/` so each has exactly one
`<!-- scaffolded by uf vdev -->` line after frontmatter.

- [x] 3.1 Clean `internal/scaffold/assets/opencode/command/`
  (6 files: cobalt-crush.md, constitution-check.md,
  finale.md, review-council.md, uf-init.md, unleash.md)
- [x] 3.2 Clean `internal/scaffold/assets/opencode/uf/packs/`
  (5 files: content.md, default.md, go.md, severity.md,
  typescript.md)
- [x] 3.3 Clean
  `internal/scaffold/assets/opencode/skill/speckit-workflow/SKILL.md`

## 4. Clean up live files

Strip duplicate markers from all live files so each has
exactly one `<!-- scaffolded by uf vdev -->` line after
frontmatter.

- [x] 4.1 Clean `.opencode/command/` embedded counterparts
  (6 files matching task 3.1)
- [x] 4.2 Clean `.opencode/uf/packs/` files
  (5 files matching task 3.2)
- [x] 4.3 Clean `.opencode/skill/speckit-workflow/SKILL.md`
- [x] 4.4 Clean non-embedded speckit commands
  (9 files: speckit.analyze.md, speckit.checklist.md,
  speckit.clarify.md, speckit.constitution.md,
  speckit.implement.md, speckit.plan.md, speckit.specify.md,
  speckit.tasks.md, speckit.taskstoissues.md)

## 5. Update tests

- [x] 5.1 Add `TestStripExistingMarkers` to
  `scaffold_test.go` covering: no markers, single marker,
  multiple markers, mixed HTML/hash markers, content
  preservation, frontmatter preservation.
- [x] 5.2 Update the `"double insert on repeat call"` test
  case in `TestInsertMarkerAfterFrontmatter`. Change
  expected output from double-marker to single-marker
  (function is now idempotent).
- [x] 5.3 Add `TestEmbeddedAssets_SingleMarker` regression
  test: walk all embedded `.md` assets and assert each
  contains at most one line matching the scaffold marker
  pattern.

## 6. Verification

- [x] 6.1 Run `go test -race -count=1 ./...` and confirm
  all tests pass (including drift detection).
- [x] 6.2 Run `go vet ./...` and confirm no issues.
- [x] 6.3 Run `golangci-lint run` and confirm no issues.
- [x] 6.4 Verify no file in the repo contains more than one
  scaffold marker line (grep check).
- [x] 6.5 Constitution alignment: Observable Quality (III)
  -- provenance markers are correct (single, accurate
  version); Testability (IV) -- new pure functions have
  isolation tests.
