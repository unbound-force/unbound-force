# Security Risk Register

Known risks for the council review pipeline. Items marked
**Accepted** have mitigations that reduce severity to an acceptable
level. Items marked **Tracked** have open issues for further
hardening.

## Tracked Risks (open issues)

### T1: Automatic invocation enables token exhaustion

**Risk**: Any PR triggers a review, consuming Vertex AI tokens.
A flood of PRs or rapid re-pushes could exhaust the token budget.

**Mitigations in place**: Org membership gate (non-members skipped),
concurrency dedup (cancel in-progress on new push).

**Tracked in**: complytime/org-infra#429 (comment trigger),
complytime/org-infra#430 (token consumption controls).

### T2: Network exfiltration via model tool use

**Risk**: A prompt injection could cause the model to exfiltrate
private repo data via shell commands, web fetches, or MCP servers.

**Mitigations in place**:
- `OPENCODE_CONFIG_CONTENT` permission config denies bash, edit,
  webfetch, websearch, skill at the OpenCode runtime level
  (hard enforcement via the permission system)
- `--pure` flag prevents external MCP plugins from loading
- Agent frontmatter (`permission: bash: deny, edit: deny,
  webfetch: deny`) provides defense-in-depth at the agent level
- Prompt instructions prohibit shell/network use (soft boundary)

**Tracked in**: complytime/org-infra#429 Phase 2 (harden-runner
egress block — network-level enforcement, independent of OpenCode).

### T3: Workflow file tampering

**Risk**: A PR could modify the council review workflow files to
weaken security controls.

**Mitigations in place**: `issue_comment` trigger (once implemented)
runs workflow code from `main`, not the PR branch.

**Tracked in**: complytime/org-infra#429 Phase 4 (CODEOWNERS), Phase
5 (workflow change detection).

## Accepted Risks

### A1: Output content injection

**Risk**: A successful prompt injection could craft the model's review
output to contain misleading security advice, phishing links, or
GitHub markdown that mimics trusted UI elements (e.g., fake approval
badges). The review output is posted directly to the PR with no
sanitization between model output and `gh pr comment`.

**Mitigations**: The prompt's defensive preamble instructs the model
to treat all diff content as untrusted and ignore override attempts.
The output is constrained to a fixed JSON schema (summary string +
inline comments array). The `filter-diff-lines.py` script validates
inline comments against diff hunks. Review comments are visibly
attributed to `github-actions[bot]`, not a human reviewer.

**Residual risk**: Low. The model would need to break out of its JSON
output contract AND the structured parsing. Even if successful, the
output is clearly bot-attributed. Reviewers should treat bot comments
with the same scrutiny as any automated tool output.

### A2: Existing review comments as injection vector

**Risk**: `prefetch.sh` passes full review bodies and inline comments
from human reviewers to the model as context. A malicious reviewer
(or compromised account) could plant prompt override instructions in
a review comment that the model then follows.

**Mitigations**: The prompt's defensive preamble applies to all input
including pre-fetched context. The org membership gate limits who can
open PRs that trigger reviews. Reviewer access requires org
membership or collaborator status.

**Residual risk**: Low. Requires a compromised org member account AND
a successful prompt injection through the defensive preamble. The
blast radius is limited to the review output — the model cannot
execute commands (the `OPENCODE_CONFIG_CONTENT` permission config
enforces this at the OpenCode runtime level).

### A3: Secrets in process environment

**Risk**: The OpenCode process runs with WIF credentials
(`GOOGLE_APPLICATION_CREDENTIALS`) and GCP project configuration
in its environment. If the model finds a way to read process
environment through a non-denied tool path, secrets could leak
into the review output posted publicly.

**Current state**: `GH_TOKEN` is unset before the `opencode run`
invocation (`run-review.sh`). `GOOGLE_APPLICATION_CREDENTIALS`
**cannot** be unset — the Vertex AI SDK reads this env var to
locate the WIF credential file for API authentication. Unsetting
it breaks auth entirely (discovered via complytime/org-infra run
#30445719456 where the review silently failed with "Non-JSON
output — comment fallback" in ~3s).

**Defense-in-depth options evaluated**:

| Approach | Security | Feasibility | Status |
|----------|----------|-------------|--------|
| Unset `GOOGLE_APPLICATION_CREDENTIALS` | Breaks auth | Not viable | Rejected |
| `OPENCODE_CONFIG_CONTENT` permission config | Denies bash/edit/webfetch/websearch/skill at runtime | Env var, merged with provider config | **Implemented** |
| `--pure` flag (no external plugins) | Reduces tool surface | CLI flag | **Implemented** |
| Agent frontmatter (`permission:`) | Denies edit/bash/webfetch per agent | Already in place | **Implemented** |
| Authenticated reverse proxy | Best — credentials never in OpenCode's process | Heavy to implement for CI | Deferred |
| Move credential file outside workspace | Marginal — SDK reads file on behalf of process | Fragile, doesn't change threat model | Not pursued |

**Mitigations in place**:
- `OPENCODE_CONFIG_CONTENT` permission config blocks shell-based
  access to process environment (`env`, `printenv`, arbitrary bash)
  at the OpenCode runtime level (bash: deny). The `read` tool
  remains allowed but `external_directory: deny` confines it to
  the project workspace; on GitHub-hosted runners the WIF
  credential file resides in `RUNNER_TEMP` outside the workspace
- `--pure` prevents external MCP plugins from bypassing permission
  restrictions
- `GH_TOKEN` is unset so GitHub API access is removed
- Vertex AI WIF tokens are short-lived (1 hour) and scoped to the
  specific project
- Prompt instructions prohibit shell and environment access
  (defense-in-depth, soft boundary)

**Residual risk**: Low. The model has no tool path to read process
environment — bash is denied at the runtime level, not just by
prompt instruction. A vulnerability in OpenCode's permission
enforcement could bypass this, but that would be an OpenCode bug.

### A4: Artifact integrity between workflows

**Risk**: The collect workflow uploads `pr-diff.patch` and
`pr-meta.json` as an artifact. The consumer downloads it. A
compromised or concurrent workflow could theoretically swap the
artifact contents.

**Mitigations**: GitHub scopes artifacts to the specific workflow run
ID. The consumer downloads using `run-id: ${{ inputs.triggering-run-id }}`
which is passed from the `workflow_run` event. Artifact substitution
would require compromising the GitHub Actions infrastructure itself.

**Residual risk**: Very low. GitHub's artifact isolation is a platform
guarantee.

### A5: Linked issue body injection

**Risk**: `prefetch.sh` fetches linked issue bodies (from
`Fixes #N` / `Closes #N` in the PR description). A malicious issue
body could contain prompt override instructions.

**Mitigations**: Same defensive preamble as A2. Issue creation in
the org repos is limited to members. The PR title (used to reference
issues) is truncated to 200 characters.

**Residual risk**: Low. Same reasoning as A2 — requires compromised
org member account plus successful prompt injection.

### A6: Log exposure of private repo content

**Risk**: Workflow logs are visible to anyone with repo read access.
The review output (including code snippets from private repos)
appears in logs via `echo` and `gh` commands.

**Mitigations**: For private repos, log access requires the same repo
read permission as code access — no escalation. For public repos,
the code is already public. The review output summarizes findings
rather than echoing full file contents.

**Residual risk**: Very low. No privilege escalation — log readers
already have code access.

### A7: npm supply chain for OpenCode CLI

**Risk**: `opencode-ai@1.15.13` is installed via `npm install -g`
which verifies package integrity via the npm registry's `integrity`
field (SHA-512 shasum). This is name-addressed, not content-addressed
— if the npm registry were compromised or the package republished
under the same version, the install would succeed with different
content. The OpenCode CLI is the primary tool processing untrusted
input.

**Mitigations**: The version is pinned to `1.15.13` (not a range or
`latest` tag). npm's built-in integrity verification checks the
SHA-512 shasum from the registry on every install. GitHub-hosted
runners use npm from the official Node.js distribution. The OpenCode
CLI runs with restricted permissions (`OPENCODE_CONFIG_CONTENT`
denies bash/edit/webfetch/websearch/skill) limiting blast radius
even if the package were compromised.

**Residual risk**: Low. An npm registry compromise affecting a
specific pinned version is a supply chain event with broad ecosystem
impact, not specific to this action. The permission sandbox limits
what a compromised CLI could do.
