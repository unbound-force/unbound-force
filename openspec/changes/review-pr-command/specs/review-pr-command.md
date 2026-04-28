## Overview

Add a standardized `/review-pr` command to the `uf init`
scaffold that reviews GitHub pull requests for CI status,
code quality, security, spec alignment, and constitution
compliance.

## Functional Requirements

- **FR-001** [MUST] The command MUST accept an optional PR
  number argument. When provided, review that specific PR.
  When omitted, detect the current branch's open PR via
  `gh pr view`.

- **FR-002** [MUST] The command MUST fetch CI check results
  and classify each failing check as PR-caused or
  pre-existing by comparing against the base branch.

- **FR-003** [MUST] The command MUST detect and run local
  deterministic tools (lint, test, format) based on project
  configuration files. If CI already ran and passed the
  same checks, local re-execution MUST be skipped.

- **FR-004** [MUST] The command MUST fetch the PR diff in a
  token-conscious manner: metadata first, diff only when
  needed. Large diffs (500+ lines) MUST be processed
  file-by-file. Binary files, lock files (`go.sum`,
  `package-lock.json`, `yarn.lock`, `bun.lock`), and
  auto-generated files (`*.pb.go`, `vendor/` contents)
  MUST be skipped. For very large PRs (2000+ lines or
  50+ files), the command MUST warn the user and offer
  to focus on specific files.

- **FR-005** [MUST] The command MUST search for associated
  specifications in `specs/`, `openspec/specs/`, and
  `openspec/changes/` directories. For Speckit specs,
  read only Functional Requirements and User Stories
  sections. For OpenSpec proposals, read only the
  Capabilities and Impact sections. If no spec is found,
  use the PR title and description as the intent source.

- **FR-006** [MUST] The AI review MUST focus on alignment
  (scope, requirement coverage, drift detection), security
  (input sanitization, injection, privilege escalation,
  secrets), and constitution compliance.

- **FR-007** [MUST] Constitution compliance checking MUST
  read `.specify/memory/constitution.md` and evaluate
  against all principles found therein. The command MUST
  NOT hardcode specific principle names or numbers.

- **FR-008** [SHOULD] When `.opencode/uf/packs/` exists,
  the command SHOULD load applicable convention packs
  (default.md, language-specific pack, custom packs) and
  reference their numbered rules in findings.

- **FR-009** [SHOULD] When `.opencode/uf/packs/severity.md`
  exists, the command SHOULD use its severity definitions.
  When absent, inline fallback definitions MUST be used.

- **FR-010** [MUST] The output MUST use a structured format
  with sections: CI Status (table), Local Tool Results,
  Summary, Alignment, Security, Constitution Compliance,
  CI Failures (PR-caused), CI Failures (Pre-existing),
  and Verdict (APPROVE / REQUEST CHANGES / COMMENT).

- **FR-011** [MUST] For pre-existing CI failures, the
  command MUST offer to create a local fix branch. Before
  creating the branch, the command MUST verify the working
  tree is clean (`git status --porcelain` is empty); if
  dirty, inform the user and skip branch creation. Fix
  branches MUST use the naming pattern
  `fix/pr-<N>-<check-name>` (e.g., `fix/pr-42-yamllint`).
  If the branch already exists, the command MUST inform
  the user and offer to switch to it or abort. The fix
  branch MUST NOT be pushed automatically. The user MUST
  confirm before branch creation. Non-trivial fixes (those
  requiring business logic changes or modifying more than
  3 files) MUST be deferred to the human with an
  explanation.

- **FR-012** [MUST] For HIGH+ findings, the command MUST
  offer to post in-line PR comments. Comments MUST be
  posted using `gh pr review <N> --comment --body <text>`
  for summary comments or `gh api` with the
  `repos/{owner}/{repo}/pulls/<N>/reviews` endpoint for
  line-specific comments. Comment body text MUST be
  shell-safe quoted to prevent injection. The command MUST
  NOT post more than 15 comments per review. All comments
  MUST be shown to the user for explicit confirmation
  before posting. The command MUST NEVER post comments
  without human approval.

- **FR-013** [MUST] The command MUST be deployed as a
  tool-owned scaffold asset at
  `internal/scaffold/assets/opencode/command/review-pr.md`.

- **FR-014** [MUST] The command file MUST use kebab-case
  naming (`review-pr.md`) consistent with all other UF
  scaffold commands.

- **FR-015** [MUST] The `gh` CLI MUST be the only external
  dependency for GitHub interaction. The command MUST
  verify `gh` availability via PATH lookup and
  authentication via `gh auth status` before making API
  calls. If `gh` is not available, error with guidance to
  install it. If not authenticated, error with guidance to
  run `gh auth login`. If authenticated but lacking
  permissions for comment posting, inform the user and
  skip the comment posting step.

## Acceptance Scenarios

### SC-001: Review PR by number

Given a repository with `uf init` scaffolding
When the user runs `/review-pr 42`
Then the command fetches PR #42 metadata, CI status, and
  diff, performs the review, and outputs structured findings
  with a verdict.

### SC-002: Auto-detect PR from branch

Given a developer on branch `feat/add-auth` with an open PR
When the user runs `/review-pr` (no argument)
Then the command detects the open PR for the current branch
  and reviews it.

### SC-003: No open PR for branch

Given a developer on a branch with no open PR
When the user runs `/review-pr` (no argument)
Then the command outputs: "No open PR found for branch
  '<branch>'. Provide a PR number: /review-pr 42"

### SC-003a: gh CLI not authenticated

Given a developer with `gh` installed but not authenticated
When the user runs `/review-pr 42`
Then the command outputs an error with guidance to run
  `gh auth login` and does not proceed with the review.

### SC-004: Convention packs enhance review

Given a repository with `.opencode/uf/packs/` containing
  `default.md` and `go.md`
When the command performs AI review
Then findings reference specific pack rules (e.g.,
  "violates CS-004") alongside constitution principles.

### SC-005: No convention packs graceful degradation

Given a repository without `.opencode/uf/packs/`
When the command performs AI review
Then the review proceeds using the constitution only,
  with no errors about missing packs.

### SC-006: Pre-existing CI failure handling

Given a PR where check `lint` fails on both the PR and the
  base branch
When the command classifies CI failures
Then the failure is classified as "Pre-existing", excluded
  from the verdict, and a fix-branch is offered after
  user confirmation.

### SC-006a: Fix-branch with dirty working tree

Given a PR with pre-existing CI failures and uncommitted
  changes in the working tree
When the command offers to create a fix branch
Then the command informs the user that the working tree is
  dirty and skips branch creation with guidance to commit
  or stash first.

### SC-007: In-line comment confirmation gate

Given the review produced 2 HIGH findings
When the command offers to post in-line comments
Then all comments are shown to the user with exact content
  and the command waits for explicit "yes" before posting.

### SC-008: Scaffold deployment

Given a fresh repository
When the user runs `uf init`
Then `.opencode/command/review-pr.md` is created as a
  tool-owned file with a version marker.

### SC-009: Scaffold update on re-run

Given a repository where `review-pr.md` was previously
  deployed and the upstream version has changed
When the user runs `uf init` again
Then the file is automatically updated to the new version
  (tool-owned behavior).
