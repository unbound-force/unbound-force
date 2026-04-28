## 1. Create Standardized Command File

- [x] 1.1 Create `.opencode/command/review-pr.md` as the
  canonical source (live copy). Adapt from org-infra's
  `review_pr.md` with all standardization deltas:
  - Kebab-case filename (`review-pr.md`)
  - Optional PR number with auto-detection preamble
    via `gh pr view --json number` (FR-001)
  - Generic constitution reference (no hardcoded
    principle names) (FR-007)
  - Convention pack awareness with graceful degradation
    (FR-008)
  - Severity pack reference with inline fallback (FR-009)
  - Proper YAML frontmatter with description
  - AI review focus areas: alignment (scope, coverage,
    drift), security (injection, escalation, secrets),
    constitution compliance (FR-006)
  - Structured output sections: CI Status table, Local
    Tool Results, Summary, Alignment, Security,
    Constitution Compliance, CI Failures (PR-caused),
    CI Failures (Pre-existing), Verdict (FR-010)
  - `gh` CLI prerequisite check via `gh auth status`
    with actionable error messages (FR-015)
  - Auto-generated file skip list: `go.sum`,
    `package-lock.json`, `yarn.lock`, `bun.lock`,
    `*.pb.go`, `vendor/` (FR-004)
  - Spec search across `specs/`, `openspec/specs/`,
    and `openspec/changes/` with format-aware section
    reading (FR-005)
  - Fix-branch guardrails: dirty-tree check, user
    confirmation, `fix/pr-<N>-<check>` naming,
    collision handling, non-trivial threshold (FR-011)
  - In-line comment guardrails: explicit `gh` subcommand
    usage, shell-safe quoting, 15-comment cap,
    human confirmation gate (FR-012)
  - All original operational steps preserved (CI
    causality, local tools, scoped diff)
  - Content packs (`content.md`, `content-custom.md`)
    excluded from detection — they contain writing
    standards, not code quality rules

## 2. Scaffold Asset Integration

- [x] 2.1 Copy `.opencode/command/review-pr.md` to
  `internal/scaffold/assets/opencode/command/review-pr.md`
  (embedded scaffold asset copy).
- [x] 2.2 Update `expectedAssetPaths` in
  `internal/scaffold/scaffold_test.go` — add
  `"opencode/command/review-pr.md"` to the slice
  (alphabetically after `review-council.md`). Update
  the `// OpenCode commands (7)` comment to `(8)`.

## 3. Documentation

- [x] 3.1 Update AGENTS.md — add `/review-pr` to the
  Project Structure tree under `.opencode/command/`.
- [x] 3.2 Update AGENTS.md — add a "PR Review" section or
  table documenting when to use `/review-pr` vs
  `/review-council`, including the command's capabilities
  and `gh` CLI prerequisite.
- [x] 3.3 Assess Website Documentation Gate — determine
  whether `/review-pr` requires a GitHub issue in
  `unbound-force/website`. If so, create it via
  `gh issue create --repo unbound-force/website`.

## 4. Verification

- [x] 4.1 Run `make check` — verify build, test, vet, and
  lint all pass with the new asset file.
- [x] 4.2 Verify `TestAssetPaths_MatchExpected` passes
  (confirms `expectedAssetPaths` matches actual embedded
  files).
- [x] 4.3 Verify `TestEmbeddedAssets_MatchSource` passes
  (confirms live copy and scaffold asset are byte-identical
  after version marker insertion). Depends on both task
  1.1 (canonical source) and task 2.1 (scaffold asset).
- [x] 4.4 Verify `TestRun_CreatesFiles` passes (confirms
  file count includes the new asset — uses
  `len(expectedAssetPaths)` dynamically).
- [x] 4.5 Verify `TestIsToolOwned` recognizes the new
  command as tool-owned (no code change needed — all
  `opencode/command/` files are tool-owned by prefix
  match).
- [x] 4.6 Verify `TestCanonicalSources_AreEmbedded` passes
  (confirms the new canonical source at
  `.opencode/command/review-pr.md` is tracked in
  `expectedAssetPaths`).
- [x] 4.7 Verify constitution alignment: Composability
  (command works standalone without other UF tools),
  Observable Quality (structured output with severity
  levels), Testability (scaffold test suite covers
  deployment).
<!-- spec-review: passed -->
<!-- code-review: passed -->
