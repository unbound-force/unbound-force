<!--
  [P] marks tasks eligible for parallel execution.
  Add [P] when a task: (a) touches different files from
  other [P] tasks in the group, (b) has no dependency
  on prior tasks in the group, (c) can safely execute
  without ordering constraints.
  Do NOT add [P] when tasks modify the same file —
  parallel workers will cause merge conflicts.
  Tasks without [P] run sequentially first, then [P]
  tasks run in parallel.
-->

## 1. Local CI gating (ci_local.yml)

- [x] 1.1 Add a `changes` job to `.github/workflows/ci_local.yml`:
      checkout with `fetch-depth: 0` (the workflow's existing
      SHA-pinned `actions/checkout`), then for `pull_request`
      events run `git diff -z --no-renames --raw` between the
      pull request base and head and set the `code` output
      per CI-FR-001 — `'true'` when any changed entry is
      outside `.uf/dewey/learnings/`, not a `.md` file, or
      not a regular file (modes `100644`/`100755`),
      `'false'` only when every changed entry is a regular
      `.md` file under it; iterate entries NUL-delimited
      (`IFS= read -r -d ''`), match the prefix and suffix
      literally against each raw path, and write `code`
      through a delimiter-based `GITHUB_OUTPUT` heredoc;
      events that are not `pull_request` and any
      indeterminate state set `'true'` without diffing.
- [x] 1.2 Give `build-and-test`, `plugin-tests`, and
      `council-review-action-tests` `needs: [changes]` and
      `if: needs.changes.outputs.code == 'true'`; leave every
      step, command, threshold, and pin untouched (CI-FR-002).
- [x] 1.3 Add the `ci-ok` aggregate job (`name: CI OK`,
      `if: always()`, `needs` the `changes` job and all three
      gated jobs, failing when any needed job concluded
      anything other than `success` or `skipped`) per
      CI-FR-003 and design Decision 4.
- [x] 1.4 Checkpoint: run `make check` and confirm the full
      suite passes (the workflow YAML itself is validated by
      the Standardized CI MegaLinter run on the implementing
      PR).

## 2. Standardized CI gating (ci_checks.yml)

- [x] 2.1 Add the same `changes` job to
      `.github/workflows/ci_checks.yml`, with logic identical
      to task 1.1 (CI-FR-001 pins both copies to the same
      behavior).
- [x] 2.2 Give `call_reusable_ci` `needs: [changes]` and the
      same gating `if:` (CI-FR-002).
- [x] 2.3 Add the aggregate job (`name: Standardized CI OK`,
      `if: always()`, `needs` the `changes` job and
      `call_reusable_ci`, failing on any conclusion other
      than `success` or `skipped`) per CI-FR-003.
- [x] 2.4 Checkpoint: run `make check` and confirm the full
      suite passes.

## 3. Required-checks ruleset update

- [x] 3.1 List the repository rulesets (`gh api
      repos/unbound-force/unbound-force/rulesets`) and
      identify the required-checks ruleset covering
      `Build and Test`, `Plugin Tests`, and
      `Standardized CI / Run linters`; record its id.
- [ ] 3.2 After the workflow change merges (design Decision 5
      ordering), PATCH that ruleset via `gh api` so the
      required checks are `CI OK` and `Standardized CI OK`,
      removing the three previous names.
- [ ] 3.3 Read the ruleset back with `gh api` and verify the
      required-check names and that no learnings PR is left
      waiting on an unreported check.
- [ ] 3.4 Exercise the CI-FR-003 failure path: if the ruleset
      PATCH fails after the workflow change merges, surface
      the failure, retry the swap until the read-back
      confirms both aggregate checks, let no pull request
      merge on an unreported required check during the
      unverified window, and if verification does not
      succeed, revert the workflow change or restore the
      required checks to `Build and Test`, `Plugin Tests`,
      and `Standardized CI / Run linters`
      (specs/ci/spec.md:166-177, scenario :216-223).

## 4. Verification and documentation

- [ ] 4.1 On the implementing PR (which changes workflow
      files, so `code` is `'true'`), confirm the `changes` job
      classifies it as code present and that every gated job
      and both aggregates run and pass (CI-FR-002,
      CI-FR-003).
- [ ] 4.2 After merge, on the next learnings PR
      (`chore/dewey-learnings-*`), confirm the gated jobs
      skip, both aggregates pass, and required checks complete
      within a couple of minutes without running the build,
      tests, or linters (CI-FR-001..CI-FR-003), while the
      security, dependency, and CRAP scans still run
      (CI-FR-004).
- [x] 4.3 Add a change entry to `CHANGELOG.md` following its
      existing format.
- [x] 4.4 Verify the implementation's constitution alignment
      against `proposal.md`: no new third-party actions or
      pins (Security by Default), fail-safe detection
      direction (gate coverage preserved), the aggregates
      always report a verdict (Observable Quality), and push
      runs on `main` are unchanged.
- [x] 4.5 Before the implementing PR merges, file the
      documentation issue for this governance-affecting CI
      change against unbound-force/unbound-force, covering the
      required-checks swap to `CI OK` and `Standardized CI OK`
      and the learnings-only skip rule, and referencing
      originating issue #1194 and the change name
      ci-skip-learnings-only-prs for traceability
      (Documentation gate: filed deliberately although
      CI-only changes are exempt, because the change alters
      user-visible merge governance).
- [x] 4.6 Verify the five adversarial CI-FR-001 scenarios
      (specs/ci/spec.md:58-111) against the detection block
      in BOTH workflows (`ci_local.yml` and `ci_checks.yml`)
      via a local dry-run of the detection logic or a test
      PR: a renamed code file (`src/auth.go` moved to
      `.uf/dewey/learnings/auth.md`), a symlink (git mode
      `120000`), a non-Markdown file under the prefix, a
      crafted filename (embedded newline, quote, or shell
      metacharacter), and an indeterminate diff — each
      classifying `code='true'` (fail-safe); this executes
      the scenarios tasks 4.1 and 4.2 do not cover, so
      design Decision 6's scenario-pinning rationale holds
      for both copies.
