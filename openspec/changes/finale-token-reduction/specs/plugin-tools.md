## ADDED Requirements

### Requirement: normalize_model_name tool

The `uf-workflow` plugin MUST provide a `normalize_model_name`
tool that accepts a raw model identifier string and returns
the normalized model family name.

The tool MUST implement the following algorithm:
1. Strip everything before and including the last `/`
2. Strip everything after and including the first `@`
3. Strip a trailing date suffix matching `-YYYYMMDD`
4. Repeatedly strip trailing version segments matching `-N`
   (a hyphen followed by a single digit at the end)
5. Validate the result contains only `[a-zA-Z0-9._-]`
6. Return `unknown-model` if the result is empty or invalid

The tool MUST return `{normalized: string}` on success.
The tool MUST return `{normalized: "unknown-model", warning:
string}` when the input cannot be normalized.

#### Scenario: Standard model identifier

- **GIVEN** a raw model identifier
  `google-vertex-anthropic/claude-sonnet-4-20250514@default`
- **WHEN** `normalize_model_name` is called
- **THEN** the tool returns `{normalized: "claude-sonnet-4"}`

#### Scenario: Model without provider prefix

- **GIVEN** a raw model identifier `claude-sonnet-4`
- **WHEN** `normalize_model_name` is called
- **THEN** the tool returns `{normalized: "claude-sonnet-4"}`

#### Scenario: Empty or invalid input

- **GIVEN** a raw model identifier that is empty or
  contains only special characters
- **WHEN** `normalize_model_name` is called
- **THEN** the tool returns
  `{normalized: "unknown-model", warning: "..."}`

---

### Requirement: Git operation tools

The `uf-workflow` plugin MUST provide the following git
operation tools. Each tool MUST return minimal structured
JSON to the LLM. (Superseded by `fix-tui-git-output-leak`,
openspec/changes/fix-tui-git-output-leak/: tool handlers
MUST NOT print full git output to TUI via `console.log()`;
all output flows through the JSON result envelope.)

All git tools MUST accept a `GitExec` function for
testability.

#### git_fetch

The tool MUST accept `remote` (string) and `branch`
(optional string) parameters.

The tool MUST return `{success: true}` on success or a
failure envelope on error.

#### Scenario: Fetch remote branch

- **GIVEN** a valid remote name `origin` and branch `main`
- **WHEN** `git_fetch` is called
- **THEN** the tool executes `git fetch origin main` and
  returns `{success: true}`

#### git_push

The tool MUST accept optional `setUpstream` (boolean) and
`branch` (optional string) parameters.

The tool MUST return `{success: true}` on success or a
failure envelope with the error message.

#### Scenario: Push with set-upstream

- **GIVEN** `setUpstream: true` and `branch: "feat/x"`
- **WHEN** `git_push` is called
- **THEN** the tool executes
  `git push --set-upstream origin feat/x`
  (Superseded by `retroactive-push-skill-contracts`,
  openspec/changes/retroactive-push-skill-contracts/: `branch`
  is required and MUST be the currently checked-out branch;
  the push uses an explicit refspec
  `HEAD:refs/heads/<branch>`, `remote` is an optional git
  config nickname, and the result is structured `GitPushData`
  with an optional `warning` when tracking configuration
  fails after a landed push. That change MUST archive after
  this change.)

#### git_push_force_with_lease

The tool MUST execute `git push --force-with-lease` and
return `{success: true}` or a failure envelope.

#### Scenario: Force push with lease

- **GIVEN** a branch with diverged history
- **WHEN** `git_push_force_with_lease` is called
- **THEN** the tool executes `git push --force-with-lease`
  (Superseded by `retroactive-push-skill-contracts`,
  openspec/changes/retroactive-push-skill-contracts/: `branch`
  is required, `remote` and `expected_sha` are optional, the
  lease binds to the expected SHA (else the remote-tracking
  ref, else absence), and the result is structured
  `GitPushData`. That change MUST archive after this change.)

#### git_merge

The tool MUST accept a `ref` (string) parameter.

The tool MUST return
`{success: true, conflicted: false}` on clean merge or
`{success: false, conflicted: true, conflictingFiles: [...]}`.

#### Scenario: Clean merge

- **GIVEN** a ref `origin/main` that merges cleanly
- **WHEN** `git_merge` is called with `ref: "origin/main"`
- **THEN** the tool returns
  `{success: true, conflicted: false}`

#### Scenario: Merge with conflicts

- **GIVEN** a ref `origin/main` that conflicts
- **WHEN** `git_merge` is called
- **THEN** the tool returns `{success: false,
  conflicted: true, conflictingFiles: ["file.go"]}`

#### git_merge_abort

The tool MUST execute `git merge --abort` and return
`{success: true}`.

#### git_rebase

The tool MUST accept a `ref` (string) parameter.

The tool MUST return
`{success: true, conflicted: false}` on clean rebase or
`{success: false, conflicted: true, conflictingFiles: [...]}`.

#### git_rebase_abort

The tool MUST execute `git rebase --abort` and return
`{success: true}`.

#### git_conflict_files

The tool MUST return `{files: [...]}` listing all files
with unresolved merge conflicts.

#### git_commit_no_edit

The tool MUST execute `git commit --no-edit` and return
`{success: true, sha: "..."}` or a failure envelope.

---

### Requirement: uf_finale_startup tool

The `uf-finale` plugin MUST provide `uf_finale_startup`
that gathers branch safety and working tree context.

The tool MUST return:
- `branch`: current branch name
- `isMain`: boolean, true if on main/master
- `branchType`: `"opsx"`, `"speckit"`, or `"other"`
- `changeName`: extracted change name from branch
- `hasChanges`: boolean, true if working tree has changes
- `secretCandidates`: array of filenames matching secret
  patterns (`.env`, `*.pem`, `*.key`, `id_rsa`, etc.)
- `unpushedCommits`: array of `{sha, subject}` for commits
  not yet pushed

#### Scenario: Feature branch with changes

- **GIVEN** the current branch is `opsx/finale-token-reduction`
  with uncommitted changes and no secret files
- **WHEN** `uf_finale_startup` is called
- **THEN** the tool returns `{branch:
  "opsx/finale-token-reduction", isMain: false,
  branchType: "opsx",
  changeName: "finale-token-reduction",
  hasChanges: true, secretCandidates: [],
  unpushedCommits: [...]}`

#### Scenario: Main branch

- **GIVEN** the current branch is `main`
- **WHEN** `uf_finale_startup` is called
- **THEN** the tool returns `{isMain: true, ...}`

#### Scenario: Secret candidates detected

- **GIVEN** the working tree contains a file named
  `.env.local` and `server.key`
- **WHEN** `uf_finale_startup` is called
- **THEN** the tool returns `{..., secretCandidates:
  [".env.local", "server.key"]}` for the LLM to
  present as a blocking gate

---

### Requirement: uf_finale_stage tool

The `uf-finale` plugin MUST provide `uf_finale_stage`
that stages all changes and returns the file count.

**Governance reconciliation**: AGENTS.md mandates "Never
use `git add -A` or `git add .` on feature branches --
stage files explicitly." This tool uses `git add .` as a
programmatic convenience ONLY because it is gated behind
a mandatory pre-staging human confirmation step: the
rewritten `/uf.finale` command MUST present the full
file list via the `question` tool and obtain explicit
user approval BEFORE calling `uf_finale_stage`. The
command template MUST NOT call this tool without prior
human confirmation. (Superseded by `fix-tui-git-output-leak`,
openspec/changes/fix-tui-git-output-leak/: the tool MUST NOT
log the file list to TUI; the staged file list is returned
in the JSON result envelope as `files: string[]`.)

The tool MUST return `{fileCount: N, files: [...]}` (the
`files` array was added by `fix-tui-git-output-leak`,
openspec/changes/fix-tui-git-output-leak/).

#### Scenario: Stage working tree after confirmation

- **GIVEN** the working tree has 5 modified files and
  the user has confirmed staging via the question tool
- **WHEN** `uf_finale_stage` is called
- **THEN** the tool executes `git add .` and returns
  `{fileCount: 5, files: [...]}`

---

### Requirement: uf_finale_commit_context tool

The `uf-finale` plugin MUST provide
`uf_finale_commit_context` that gathers context for commit
message generation.

The tool MUST accept `rawModelId` (string) parameter.

The tool MUST return:
- `diffStat`: output of `git diff --cached --stat`
- `recentLog`: output of `git log -5 --oneline`
- `normalizedModel`: normalized model name (via
  `normalize_model_name` algorithm)

#### Scenario: Gather commit context

- **GIVEN** staged changes and raw model identifier
  `provider/claude-sonnet-4-20250514@default`
- **WHEN** `uf_finale_commit_context` is called
- **THEN** the tool returns `{diffStat: "...",
  recentLog: "...",
  normalizedModel: "claude-sonnet-4"}`

---

### Requirement: uf_finale_push_state tool

The `uf-finale` plugin MUST provide `uf_finale_push_state`
that detects push state and divergence.

The tool MUST return:
- `upstreamExists`: boolean
- `diverged`: boolean
- `aheadBy`: number of commits ahead
- `behindBy`: number of commits behind

#### Scenario: Branch ahead of upstream

- **GIVEN** the branch has 3 unpushed commits and no
  upstream changes
- **WHEN** `uf_finale_push_state` is called
- **THEN** the tool returns `{upstreamExists: true,
  diverged: false, aheadBy: 3, behindBy: 0}`

---

### Requirement: uf_finale_pr_context tool

The `uf-finale` plugin MUST provide `uf_finale_pr_context`
that gathers PR creation context.

The tool MUST return:
- `existingPr`: `{number, url, title}` or `null`
- `isFork`: boolean
- `forkParent`: `{owner, repo}` or `null`
- `templateSections`: array of section names from
  `.github/PULL_REQUEST_TEMPLATE.md`
- `commitLog`: array of `{sha, subject}`
- `diffStat`: output of diff stat against base
- `originatingIssue`: number or `null` (from
  `.openspec.yaml` `originating_issue` field)
- `unleashSummary`: `{whatWasBuilt, howToVerify,
  testResults}` or `null` (from
  `.uf/artifacts/unleash-summary/<branch>.json`)

#### Scenario: New PR with unleash-summary

- **GIVEN** no existing PR for the branch and an
  unleash-summary artifact exists
- **WHEN** `uf_finale_pr_context` is called
- **THEN** the tool returns `{existingPr: null, ...,
  unleashSummary: {whatWasBuilt: "...",
  howToVerify: "...", testResults: "..."}}`

#### Scenario: Existing PR

- **GIVEN** an open PR #42 exists for the branch
- **WHEN** `uf_finale_pr_context` is called
- **THEN** the tool returns
  `{existingPr: {number: 42, url: "...", title: "..."},
  ...}`

---

### Requirement: uf_finale_create_pr tool

The `uf-finale` plugin MUST provide `uf_finale_create_pr`
that creates a PR via the `gh` CLI.

The tool MUST accept `title` (string), `body` (string),
and optional `targetRepo` (string) parameters.

The tool MUST implement a secure temp file lifecycle:
1. Create the temp file via `mktemp` or platform equivalent
2. Set permissions to `chmod 600` immediately after creation
3. Write the body content to the temp file
4. Pass the temp file path via `--body-file` to `gh pr create`
5. Remove the temp file via `rm -f` on ALL exit paths
   (success, failure, and exception)
6. If `mktemp` fails, the tool MUST return a failure envelope
   with code `TEMP_FILE_ERROR` — it MUST NOT fall back to
   inline `--body`

The tool MUST return `{number: N, url: "..."}`.

#### Scenario: Create PR

- **GIVEN** a title and body for a new PR
- **WHEN** `uf_finale_create_pr` is called
- **THEN** the tool creates the PR via `gh pr create` with
  `--body-file`, cleans up the temp file, and returns
  `{number: 42, url: "https://..."}`

#### Scenario: Temp file creation failure

- **GIVEN** `mktemp` fails (e.g., disk full, permissions)
- **WHEN** `uf_finale_create_pr` is called
- **THEN** the tool returns a failure envelope with
  code `TEMP_FILE_ERROR` and does NOT fall back to
  inline `--body`

---

### Requirement: uf_finale_ci_poll tool

The `uf-finale` plugin MUST provide `uf_finale_ci_poll`
that polls CI checks with an internal loop.

The tool MUST accept `prNumber` (number) and optional
`timeoutMinutes` (number, default 15) parameters.

The tool MUST poll every 30 seconds until all checks
complete or timeout.

The tool MUST return:
- `status`: one of `pass`, `fail`, `conflict`,
  `no-checks-workflows-exist`, `no-checks-no-workflows`
- `failedChecks`: array of `{name, conclusion, url}`
- `mergeable`: boolean or null
- `workflowsConfigured`: boolean
- `skipReason`: string or null (when no checks exist)

#### Scenario: All checks pass

- **GIVEN** PR #42 with 3 CI checks all passing
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool blocks until checks complete and
  returns `{status: "pass", failedChecks: [],
  mergeable: true, workflowsConfigured: true}`

#### Scenario: Timeout

- **GIVEN** PR #42 with checks still running after
  timeout
- **WHEN** `uf_finale_ci_poll` is called with
  `timeoutMinutes: 1`
- **THEN** the tool returns a failure envelope with
  code `CI_POLL_TIMEOUT`

#### Scenario: gh CLI unavailable or auth failure

- **GIVEN** the `gh` CLI is not installed, not in PATH,
  or not authenticated
- **WHEN** `uf_finale_ci_poll` is called
- **THEN** the tool returns a failure envelope with
  code `GH_CLI_ERROR` and the stderr output

---

### Requirement: uf_finale_return_to_main tool

The `uf-finale` plugin MUST provide
`uf_finale_return_to_main` that checks out main and pulls.

The tool MUST return `{success: true}`.

#### Scenario: Return to main

- **GIVEN** the current branch is `opsx/foo`
- **WHEN** `uf_finale_return_to_main` is called
- **THEN** the tool executes `git checkout main` and
  `git pull` and returns `{success: true}`

---

### Requirement: Uniform plugin probe validation

The scaffold's `reviewPluginProbeSource()` function MUST
import and validate ALL registered plugins, not just a
subset.

The probe source MUST validate at least one tool name
from each plugin:
- `invoke-agent`: `invoke_agent`
- `review-dispatch`: `resolve_base_ref`
- `uf-workflow`: `normalize_model_name`
- `uf-finale`: `uf_finale_startup`

#### Scenario: All plugins probe-validated

- **GIVEN** four plugins are registered (invoke-agent,
  review-dispatch, uf-workflow, uf-finale)
- **WHEN** the probe source is generated
- **THEN** all four plugins are imported and their
  representative tool names are checked

## MODIFIED Requirements

### Requirement: Scaffold asset inventory

The scaffold embedded asset count changes from 67 to 68
total (1 non-deployed devcontainer unchanged). Deployed
base count remains 63 (the new uf-finale asset is
activation-gated, not deployed by default).
Activation-gated plugin count changes from 3 to 4.

Previously: 67 total, 63 deployed, 3 activation-gated.
After: 68 total, 63 deployed, 4 activation-gated.

## REMOVED Requirements

None.
