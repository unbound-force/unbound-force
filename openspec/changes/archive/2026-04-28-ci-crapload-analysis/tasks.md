## 1. Dependabot Configuration

- [x] 1.1 Create `.github/dependabot.yml` with YAML version 2,
  `github-actions` ecosystem (directory `/`, daily schedule,
  `ci` commit prefix with scope), and `gomod` ecosystem
  (directory `/`, weekly schedule, 10 open PR limit, `chore`
  commit prefix with scope).

## 2. CRAP Load CI Workflow

- [x] 2.1 Create `.github/workflows/ci_crapload.yml` with
  `# SPDX-License-Identifier: Apache-2.0` header, descriptive
  comment block, `pull_request` trigger on `main`, and
  workflow-level `permissions` (contents read, pull-requests
  write).
- [x] 2.2 Add `crapload` job calling
  `complytime/org-infra/.github/workflows/reusable_crapload_analysis.yml@9205a3ac6b76b75dbe6e22b2f0f330bc8edbeb38`
  with `permissions: contents: read`.
- [x] 2.3 Add `post-comment` job with `needs: crapload`,
  `if: ${{ !cancelled() }}`, `permissions: pull-requests: write`.
  Download `crapload-analysis` artifact using pinned
  `actions/download-artifact` SHA with version comment. Post or
  update PR comment using pinned `actions/github-script` SHA
  with version comment and
  `<!-- crapload-analysis-marker -->` idempotency marker.
  Include comment truncation logic (60000 char limit) and
  fallback message when artifact is missing.

## 3. CRAP Score Baseline

- [x] 3.1 Create `.gaze/` directory.
- [x] 3.2 Generate coverage profile by running
  `go test -coverprofile=coverage.out ./...`
- [x] 3.3 Run `gaze report --format=json --coverprofile=coverage.out ./...`,
  extract `.crap` section with path normalization (strip repo
  root prefix from file paths), and write to
  `.gaze/baseline.json`.
- [x] 3.4 Verify baseline JSON structure contains `scores` array
  with per-function entries including `package`, `function`,
  `file`, `line`, `complexity`, `line_coverage`, and `crap`
  fields.

## 4. Verification

- [x] 4.1 Validate `ci_crapload.yml` YAML syntax.
- [x] 4.2 Validate `dependabot.yml` YAML syntax.
- [x] 4.3 Verify all action references in `ci_crapload.yml` use
  pinned SHAs with version comments (not `@main` or `@vN`).
- [x] 4.4 Verify the reusable workflow reference uses the pinned
  SHA `9205a3ac6b76b75dbe6e22b2f0f330bc8edbeb38`.
<!-- spec-review: passed -->
<!-- code-review: passed -->
