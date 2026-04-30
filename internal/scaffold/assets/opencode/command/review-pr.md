---
description: "Review a pull request for alignment, security, and constitution compliance"
---
<!-- scaffolded by uf vdev -->

# Review Pull Request

You are a token-efficient code reviewer. The user will provide a PR number or you will auto-detect it from the current branch. Delegate deterministic checks to local tools and CI results first, then apply AI judgment only where tools cannot reach: intent alignment, security patterns, and architectural concerns.

## Arguments

- **PR number** (optional): The pull request number to review (e.g., `42`). If omitted, the command auto-detects the open PR for the current branch.

**Argument parsing** (before any tool calls): Check the
user's message for a PR number argument. If present, set
`PR_NUMBER` to that value immediately. All subsequent steps
use `<PR_NUMBER>` — no auto-detection commands are needed
or permitted.

## Execution Steps

### 0. Prerequisites

Verify the `gh` CLI is available and authenticated before proceeding:

```bash
which gh
```

If `gh` is not found: **STOP** with error:
> "`gh` CLI is not installed. Install it from https://cli.github.com/ or via your package manager."

If `gh` is found, verify authentication:

```bash
gh auth status
```

If not authenticated: **STOP** with error:
> "`gh` is installed but not authenticated. Run `gh auth login` to authenticate."

#### Execution Mode Check

This command requires running local tools (build, test,
lint) as part of the review. Verify you can execute
commands by running a harmless probe:

```bash
echo "mode-check-ok"
```

If the probe cannot be executed (the agent runtime
returns a tool-access-denied error, or you are in plan
mode, read-only mode, or otherwise restricted from
running commands): **STOP** with message:

> "This review requires running local tools (build,
> test, lint) to verify the PR. I am currently in
> plan/read-only mode which prevents executing these
> checks. Switch to a mode that allows command
> execution (e.g., full mode / auto mode) and
> re-invoke `/review-pr <N>`."

Do NOT proceed with a partial review that skips local
tool execution. The local tool results are the
foundation of the review — without them, AI-only
findings lack verification and the review does not
meet the command's quality standard.

### 1. Resolve PR Number

**If `PR_NUMBER` was already set from the argument**: skip
this step entirely. Do NOT run `gh pr view`,
`git branch --show-current`, or any branch/PR detection
commands.

**Only if no PR number was provided**: auto-detect from
the current branch:

```bash
gh pr view --json number --jq '.number'
```

If no open PR exists for the current branch: **STOP** with error:
> "No open PR found for branch '`<branch>`'. Provide a PR number: `/review-pr 42`"

### 2. Fetch PR Metadata (Minimal)

Retrieve PR metadata first — avoid loading the full diff until needed:

```bash
gh pr view <PR_NUMBER> --json title,body,files,additions,deletions,baseRefName,headRefName,labels,milestone,commits
```

Record the PR title, description, branch name, base branch, and changed file list. **Do NOT fetch the full diff yet** — later steps determine which files need AI analysis.

### 3. Fetch CI Check Results

Retrieve the CI/CD check suite status for the PR:

```bash
gh pr checks <PR_NUMBER> --json name,state,description,link
```

Categorize each check as:
- **PASS**: Check succeeded
- **FAIL**: Check failed
- **PENDING**: Check still running
- **SKIPPED**: Check was skipped

If checks are still PENDING, inform the user and ask whether to wait or proceed with the available results.

**If all checks pass**: Record this and move to Step 4. No CI triage needed.

**If any checks fail**: Proceed to Step 3a for causality determination.

#### 3a. CI Failure Causality Determination

For each failing check, determine whether the failure is caused by the PR's changes or is a pre-existing issue on the base branch.

**Method**: Check if the same test/check also fails on the base branch:

```bash
# Get the base branch name (from Step 2 metadata, e.g., "main")
BASE_BRANCH="<baseRefName from Step 2>"

# Check the latest CI status on the base branch
# Use --jq with $ENVIRON or --arg to avoid injection from check names containing quotes
gh api repos/{owner}/{repo}/commits/${BASE_BRANCH}/check-runs \
  --jq --arg name "<FAILING_CHECK_NAME>" '.check_runs[] | select(.name == $name) | {name, conclusion}'
```

**Classification**:

| Base branch status | PR check status | Classification |
|--------------------|-----------------|----------------|
| Pass | Fail | **PR-caused** — the PR introduced the failure |
| Fail | Fail | **Pre-existing** — failure exists independently of the PR |
| No data | Fail | **Unknown** — treat as PR-caused (conservative) |

Record the classification for each failing check. This feeds into Step 8 (AI review) and Step 10 (fix-branch).

### 4. Run Local Deterministic Tools (Pre-flight)

Run the project's own tools as a rapid pre-flight check.

**Detection**: Check which tools are available by looking
for their configuration files:

```bash
test -f Makefile && echo "MAKEFILE=yes"
test -f .golangci.yml && echo "GO_LINT=yes"
test -f ruff.toml -o -f pyproject.toml && echo "PYTHON_LINT=yes"
test -f .yamllint.yml && echo "YAML_LINT=yes"
test -f .pre-commit-config.yaml && echo "PRECOMMIT=yes"
```

**CI coverage check** (mandatory before running any
tool): Build and display a coverage matrix that maps
each detected local tool to the CI check from Step 3
that covers the same verification. Display this matrix
to make the skip/run decision visible:

| Local tool | CI check that covers it | CI status | Run locally? |
|------------|------------------------|-----------|--------------|
| `go test` | e.g., "Local CI / test" | PASS/FAIL/NONE | Yes/No |
| `golangci-lint` | e.g., "CI Checks / lint" | PASS/FAIL/NONE | Yes/No |
| ... | ... | ... | ... |

Decision rules:
- CI status PASS → skip locally ("No" — CI already
  verified)
- CI status FAIL → skip locally ("No" — failure already
  captured in Step 3a, will be analyzed in Step 8d)
- CI status NONE (no matching check) → MUST run
  locally ("Yes")
- No CI checks reported at all → MUST run ALL detected
  local tools ("Yes" for every row)

**Execution**: Run only the tools marked "Yes" in the
matrix above:

| Tool detected | Command to run | What it checks |
|---------------|----------------|----------------|
| Makefile | `make lint` (or `make check`) | Project-defined lint/format/vet |
| `.golangci.yml` | `golangci-lint run ./...` | Go lint rules |
| `ruff.toml` / `pyproject.toml` | `ruff check .` | Python lint rules |
| `.yamllint.yml` | `yamllint .` | YAML lint rules |
| `.pre-commit-config.yaml` | `pre-commit run --all-files` | Pre-commit hooks |
| `go.mod` | `go test ./...` | Go tests |
| `pyproject.toml` / `setup.py` | `pytest` or `python -m pytest` | Python tests |

**Record results**: Capture tool exit codes and output.
If tools pass, skip those categories in the AI review
entirely. If tools fail, include the failure output as
context.

**If no tools are detected**: Note this and proceed to
AI-based review for all categories.

### 5. Fetch Diff (Scoped)

Now fetch the diff, being token-conscious:

```bash
gh pr diff <PR_NUMBER>
```

**Large diff handling** (500+ lines):

`gh pr diff` does not support file path filters. For
large diffs, save the output to a temp file and
navigate it with targeted reads:

1. Save the full diff once:
   ```bash
   gh pr diff <PR_NUMBER> > /tmp/pr<PR_NUMBER>.diff
   ```
   (The tool runtime auto-saves truncated output to a
   file — use that path if available instead.)

2. Find file boundaries in the saved diff:
   ```bash
   grep -n '^diff --git' /tmp/pr<PR_NUMBER>.diff
   ```
   This returns line numbers for each file's diff
   section.

3. Read specific file sections using offset/limit on
   the saved file. Skip these files entirely:
   - Lock files: `package-lock.json`, `go.sum`,
     `yarn.lock`, `bun.lock`
   - Auto-generated: `*.pb.go`, `vendor/` contents
   - Binary files
   - CRAP baselines: `.gaze/baseline.json`

4. For very large PRs (2000+ lines or 50+ files),
   warn the user and ask whether to review all files
   or focus on specific ones.

**Do NOT attempt**:
- `gh pr diff <N> -- <path>` (unsupported, will fail)
- `git show <remote>/<branch>:<path>` (PR branch may
  not be on any configured remote)
- `git fetch <remote> <branch>` (PR may come from a
  fork or push directly to PR refs)

#### Accessing full file contents from the PR branch

If you need to read a complete file from the PR branch
(not just the diff), use the GitHub API. The PR branch
may not exist on any locally configured remote:

```bash
gh api repos/{owner}/{repo}/contents/<path>?ref=<headRefName> \
  --jq '.content' | base64 -d
```

Use `<headRefName>` from the Step 2 metadata. If the
API call returns 404, 403, or empty content (files
>1 MB), fall back to reading from the saved diff file
and note in the review that full file content was
unavailable.

For accessing files on the PR branch, the agent MUST
use `gh api` exclusively. Any `git` subcommand
targeting the PR's head ref (`git show`, `git fetch`,
`git checkout`, `git diff` with remote refs) is
prohibited.

### 6. Locate Associated Specification

Search for a specification that matches this PR across all spec directories:

- Check if the PR branch name matches a spec directory:
  - `specs/<branch-name>/spec.md` (Speckit output)
  - `openspec/specs/<branch-name>/spec.md` (OpenSpec specs)
  - `openspec/changes/<branch-name>/proposal.md` (OpenSpec changes)
- Check if the PR description references a spec
- If not found locally, check the PR's changed file
  list (from Step 2 metadata) for spec artifacts. The
  spec may be introduced by the PR itself. If found
  in the changed file list, read the spec content from
  the saved diff (Step 5) rather than from the
  filesystem.
- If a Speckit spec is found, read only the **Functional Requirements** and **User Stories** sections (not the entire spec) to minimize token usage
- If an OpenSpec proposal is found, read only the **Capabilities** and **Impact** sections
- If no spec is found in any directory or in the PR's changed files, note this and use the PR title and description as the intent source

### 7. Load Convention Packs (Optional)

Check if convention packs are available for enhanced review precision:

```bash
test -d .opencode/uf/packs && echo "PACKS=yes"
```

**If packs are available**:
1. Always read `.opencode/uf/packs/default.md` (language-agnostic rules)
2. Detect language and load the appropriate pack:
   - `go.mod` exists → read `.opencode/uf/packs/go.md`
   - `tsconfig.json` or `package.json` exists → read `.opencode/uf/packs/typescript.md`
3. Read corresponding `-custom.md` files if they exist (e.g., `go-custom.md`)
4. Read `.opencode/uf/packs/severity.md` if it exists — use its severity definitions instead of the inline fallback in Step 8
5. Do NOT load `content.md` or `content-custom.md` — these contain writing standards for documentation agents, not code quality rules

Use pack rules (CS-001, AP-001, SC-001, TC-001, DR-001, etc.) alongside the constitution for more specific, actionable findings. Reference the specific rule ID in each finding.

**If packs are NOT available**: proceed without them. Use the constitution and inline severity definitions only. No error or warning needed.

### 8. AI Review (Judgment-Based Only)

Focus AI analysis exclusively on what deterministic tools and CI cannot check. Skip any category where local tools or CI already passed.

#### 8a. Alignment Check

Compare the PR intent (title + description + linked spec) against the actual code changes:

- **Scope alignment**: Do the changed files match what the spec/description says should change? Flag files modified outside the stated scope.
- **Requirement coverage**: For each requirement in the spec (if found), verify the code changes address it. Flag uncovered requirements.
- **Completeness**: Are there partial implementations that could leave the system in an inconsistent state?
- **Drift detection**: Does the code do anything NOT described in the intent/spec? Flag undocumented behavioral changes.

#### 8b. Security Review

Examine the diff for security vulnerabilities that linters cannot catch:

- **Input sanitization**: Are external inputs (user input, API parameters, file paths, environment variables, command arguments) validated before use in:
  - SQL queries (injection risk)
  - Shell commands (command injection)
  - File paths (path traversal)
  - HTML/template output (XSS)
  - YAML/JSON parsing (deserialization attacks)
- **Unexpected workflows**: Can the code be executed in an unintended order or context?
  - Missing authentication/authorization checks
  - Race conditions or TOCTOU vulnerabilities
  - State machine violations (skipping steps)
  - Error handling that exposes sensitive information
- **Privilege escalation**: Does the code grant permissions or elevate privileges without proper validation?
- **Secrets and credentials**: Are there hardcoded secrets, tokens, or API keys? Are secrets logged or exposed in error messages?
- **Dependency risks**: Are new dependencies well-maintained and from trusted sources?

#### 8c. Constitution Compliance (AI-only items)

Read `.specify/memory/constitution.md` if it exists. Extract all principles and their MUST/SHOULD rules. For each principle, check whether the PR's changes comply. **Only check items that local tools and CI did NOT already verify.**

If no constitution file exists, note this and review against general software engineering best practices. Do NOT hardcode specific principle names or numbers — each project defines its own constitution.

**Skip if already covered by local tools or CI**: naming conventions, line length, lint issues, formatting, file headers.

#### 8d. CI Failure Analysis

For each CI failure classified in Step 3a, provide analysis:

**PR-caused failures**: Include as HIGH or CRITICAL findings:
- Which check failed and what the error output says
- Which PR change likely caused the failure (map failing test to changed file/function)
- Suggested fix or direction

**Pre-existing failures**: Report separately with clear labeling:
- Confirm the failure also exists on the base branch
- Brief root cause analysis if determinable from the error output
- Note that this will be addressed in Step 10 (fix-branch offer)

### 9. Output Format

Present findings in this structured format:

```markdown
## PR Review: #<NUMBER> — <TITLE>

### CI Status
| Check | Status | Classification |
|-------|--------|----------------|
| <name> | PASS/FAIL | PR-caused / Pre-existing / N/A |

### Local Tool Results
<Table showing which tools ran, pass/fail status, and summary of failures if any>

### Summary
<1-2 sentence overview of what the PR does and overall assessment>

### Alignment
- <Finding with severity>

### Security
- <Finding with severity>

### Constitution Compliance
- <Finding with severity>

### CI Failures (PR-caused)
- <Finding with severity — only if PR-caused failures exist>

### CI Failures (Pre-existing)
- <Description — only if pre-existing failures exist>
- Note: These failures exist independently of this PR. See fix-branch offer below.

### Verdict
**<APPROVE / REQUEST CHANGES / COMMENT>**

<Brief justification. Pre-existing CI failures do NOT block the PR verdict.>
```

**Severity levels** (use `.opencode/uf/packs/severity.md` definitions if loaded in Step 7, otherwise use these defaults):
- **CRITICAL**: Must be fixed before merge (security vulnerabilities, data loss risks)
- **HIGH**: Should be fixed before merge (spec violations, missing tests for critical paths, PR-caused CI failures)
- **MEDIUM**: Recommended to fix (code quality, minor compliance issues)
- **LOW**: Optional improvements (style, naming suggestions)

If no issues are found in a category, state "No issues found."

### 10. Offer Fix-Branch for Pre-existing CI Failures

If Step 3a identified any **pre-existing** CI failures, offer to create a fix branch:

```
I identified <N> pre-existing CI failure(s) that are NOT caused by this PR:
- <check name>: <brief description of failure>

These failures also occur on the base branch (<BASE_BRANCH>).

Would you like me to create a fix branch with a proposed resolution?
I will create the branch and commit locally — you can review the changes and file a PR when ready.
```

**If the user agrees**:

1. **Verify clean working tree**:
   ```bash
   git status --porcelain
   ```
   If the output is not empty: **STOP** branch creation with message:
   > "Working tree has uncommitted changes. Commit or stash them before creating a fix branch."
   Switch back to the PR branch and continue to Step 11.

2. **Check for branch name collision**:
   ```bash
   git branch --list "fix/pr-<PR_NUMBER>-<check-name>"
   ```
   If the branch already exists, inform the user:
   > "Branch `fix/pr-<PR_NUMBER>-<check-name>` already exists. Switch to it with `git checkout fix/pr-<PR_NUMBER>-<check-name>`, or delete it first."
   Switch back to the PR branch and continue to Step 11.

3. **Sanitize the check name** for branch-name safety:
   lowercase, replace spaces and special characters with
   hyphens, strip consecutive hyphens, remove characters
   outside `[a-z0-9._-]`, truncate to 50 characters.
   Example: `"Build (ubuntu/latest)"` → `build-ubuntu-latest`.
   Also validate that `<PR_NUMBER>` is digits only.

4. **Create a fix branch** from the base branch:
   ```bash
   git checkout <BASE_BRANCH>
   git checkout -b fix/pr-<PR_NUMBER>-<sanitized-check-name>
   ```
   Branch naming: `fix/pr-<PR_NUMBER>-<sanitized-check-name>` (e.g., `fix/pr-42-yamllint`, `fix/pr-42-test-auth-timeout`)

5. **Analyze and propose the fix**: Use the CI failure output and the failing file(s) to determine the minimal change needed. Keep the scope as small as possible — fix only what is failing.

6. **Commit with Conventional Commits format**:
   Write the commit message to a temporary file to avoid
   shell injection from AI-generated description text,
   then commit using `-F`:
   ```bash
   git add <changed-files>
   git commit -s -F <temp-commit-message-file>
   ```
   The commit message file should contain:
   ```
   fix: resolve <failing-check> CI failure

   <Brief description of what was wrong and how the fix addresses it.>

   This failure was pre-existing on <BASE_BRANCH> and unrelated to PR #<PR_NUMBER>.

   Assisted-by: OpenCode (<model>)
   ```
   Remove the temp file after committing.

7. **Report to the user**:
   ```
   Fix branch created: fix/pr-<PR_NUMBER>-<check-name>

   Changes:
   - <file>: <what changed>

   The branch is local. To review and push:
     git checkout fix/pr-<PR_NUMBER>-<check-name>
     git log -1
     git push -u origin fix/pr-<PR_NUMBER>-<check-name>
   ```

8. **Switch back** to the PR branch:
   ```bash
   git checkout <PR_BRANCH>
   ```

**Guardrails**:
- The fix MUST be scoped to the specific failing check — no unrelated changes
- The agent MUST NOT push to the remote or file a PR automatically
- If the fix is non-trivial (requires understanding business logic, architectural decisions, or modifying more than 3 files), inform the user instead of attempting a fix:
  ```
  The CI failure in <check> appears to require a non-trivial fix involving <description>.
  I recommend investigating this separately rather than proposing an automated fix.
  ```

### 11. Offer In-line PR Comments

After presenting the summary, if there are findings with severity HIGH or above, offer to post them as in-line comments on the PR:

```
I found <N> findings (X CRITICAL, Y HIGH). Would you like me to post in-line comments on the PR so the author can see them in context?

I will prepare the comments and show them to you for approval before posting anything.
```

**If the user agrees**:

1. **Prepare comments**: For each finding that maps to a specific file and line range in the diff, prepare an in-line comment with:
   - The finding description
   - The severity level
   - A concrete suggestion for fixing the issue (if applicable)
   - Cap at 15 comments maximum. If more than 15 findings qualify, prioritize CRITICAL over HIGH. For remaining findings beyond the cap, include them in a single summary comment.

2. **Show all comments for human review**: Present each prepared comment in this format:
   ```
   File: <path>
   Line: <line_number>
   Body: <comment text>
   ```

3. **Wait for explicit confirmation**: Ask "Post these comments? (yes/no/edit)"
   - **yes**: Post comments using the `gh` CLI. For a summary comment, write the body to a temporary file and use `--body-file` to avoid shell injection from AI-generated text:
     ```bash
     gh pr review <PR_NUMBER> --comment --body-file <temp-summary-file>
     ```
     For in-line comments, use the GitHub API with a JSON input file:
     ```bash
     gh api repos/{owner}/{repo}/pulls/<PR_NUMBER>/reviews \
       --method POST \
       --input <json-file-with-comments>
     ```
     The JSON file should contain the review body, event type, and inline comments array.
     Always write comment payloads to temporary files rather than interpolating AI-generated text into shell arguments, to prevent shell injection. Remove temporary files after posting. If `gh api` returns a 403 or permission error, inform the user that their token lacks write permissions for PR comments and suggest re-authenticating with `gh auth login`.
   - **no**: Skip posting, the summary is sufficient
   - **edit**: Let the user modify comments before posting, then re-confirm

4. **CRITICAL RULE**: NEVER post comments without explicit human confirmation. Always show the exact content that will be posted and wait for approval.
