---
description: Run the reviewer governance council to audit codebase or spec compliance.
---
<!-- scaffolded by uf vdev -->
# Command: /review-council

## User Input

```text
$ARGUMENTS
```

## Description

Review the current codebase **or** SpecKit artifacts for compliance with the Behavioral Constraints in `AGENTS.md` using the review council. The council dynamically discovers which reviewer agents are available rather than assuming a fixed set.

## Determine Review Mode

The review mode is determined automatically by examining the
workspace state. The user can also force a mode explicitly.

### Explicit Override

If `$ARGUMENTS` contains the word **"specs"**, use
**Spec Review Mode** regardless of auto-detection.

If `$ARGUMENTS` contains the word **"code"**, use
**Code Review Mode** regardless of auto-detection.

### PR Number Argument

After removing mode keywords ("specs", "code") from
`$ARGUMENTS`, check if the remaining text contains a
positive integer. If so, validate it:

- **Digits only** (no letters, punctuation, or signs)
- **Range 1–999999**

If valid, record it as the **explicit PR number** for
use in Phase 1c (Protocol 2) and Step 7 (GitHub
posting). If invalid (non-numeric, out of range, or
negative), reject with an informational error:

> "Invalid PR number: `<value>`. Expected a positive
> integer (1–999999). Ignoring."

And proceed without a PR number.

### Auto-Detection (when no explicit override)

When no mode keyword is provided, detect the mode by
examining the current branch and workspace:

1. **Get the current branch name**:
   ```bash
   git rev-parse --abbrev-ref HEAD
   ```

2. **Get the diff against the base branch** (`main`):
   ```bash
   git diff --name-only main...HEAD
   ```
   This shows all files changed on the current branch
   relative to `main`.

3. **Classify the changed files**:
   - **Spec files**: paths under `specs/`, `openspec/`,
     `.specify/`, or files named `spec.md`, `plan.md`,
     `tasks.md`, `checklists/`, `contracts/`,
     `data-model.md`, `research.md`
   - **Code files**: everything else (`.go`, `.ts`, `.js`,
     `.py`, `go.mod`, `go.sum`, `Makefile`, `internal/`,
     `cmd/`, `.opencode/agents/`, `.opencode/commands/`,
     `.opencode/skills/`, `.opencode/uf/packs/`,
     etc.)

4. **Detect the workflow tier** from the branch name:
   - Branch matches `opsx/*`: **OpenSpec** (tactical)
   - Branch matches `NNN-*` (digits then dash): **Speckit** (strategic)
   - Branch is `main` or other: no active workflow

5. **Select mode based on classification**:

   | Condition | Mode | Rationale |
   |-----------|------|-----------|
   | Code files changed | **Code Review** | Post-implementation -- review the code |
   | Only spec files changed | **Spec Review** | Pre-implementation -- review the specs |
   | No files changed vs main | **Spec Review** | On main or fresh branch -- review specs |
   | On `main` branch | **Spec Review** | No feature branch -- review specs |

6. **Announce the detected mode**: Always tell the user
   which mode was selected and why, including the
   workflow tier:
   > "Detected **Code Review Mode** (Speckit) — found N
   > code files changed on branch `012-swarm-delegation`
   > vs `main`."
   >
   > Or: "Detected **Spec Review Mode** (OpenSpec) — only
   > spec artifacts changed on branch
   > `opsx/documentation-accuracy`."
   >
   > Use `/review-council code` or `/review-council specs`
   > to override.

---

## Discover Available Reviewers

Before entering either review mode, discover which reviewer agents are available:

1. **Read the `.opencode/agents/` directory** using the Read tool to list all entries.

2. **Filter for Divisor persona agents**: from the directory listing, select only entries whose filename starts with `divisor-` and ends with `.md` (e.g., `divisor-adversary.md`, `divisor-architect.md`). Ignore subdirectories (entries ending with `/`) and non-matching files.

3. **Extract agent names**: for each matching file, strip the `.md` extension to get the agent name (e.g., `divisor-adversary.md` → `divisor-adversary`).

4. **Guard clause**: if zero Divisor persona agents are discovered, report to the user that no `divisor-*.md` agents were found in `.opencode/agents/` and stop. Do not proceed with either review mode.

5. **Note absent personas**: compare discovered agents against the known Divisor persona roles listed in the reference table below. Any known role not discovered is noted as absent. Absent personas are **informational only** — they do not block the review.

### Known Divisor Persona Roles (Reference Table)

This table documents known Divisor persona roles and their focus areas. It is used for context when delegating to discovered agents, but the **invocation list comes solely from discovery** — not from this table.

| Agent Name | Persona | Code Review Focus | Spec Review Focus |
|---|---|---|---|
| `divisor-adversary` | The Adversary | Secrets/credentials, dependency CVEs/supply chain, error handling/resilience, path/injection safety | Completeness, testability, ambiguity, security gaps, dependency risks, cross-spec consistency |
| `divisor-architect` | The Architect | Architectural alignment, coding conventions [PACK], pattern adherence, DRY, testing conventions [PACK], documentation [PACK] | Template consistency, spec-to-plan alignment, task coverage, data model coherence, inter-spec architecture |
| `divisor-guard` | The Guard | Intent drift/plan alignment, zero-waste mandate, constitution alignment, cross-component value [PACK] | Intent fidelity, scope discipline, inter-spec consistency, status accuracy, user value, constitution alignment |
| `divisor-testing` | The Tester | Test architecture [PACK], coverage strategy, assertion depth, test isolation, regression protection, convention compliance [PACK] | Testability of requirements, test strategy coverage, fixture feasibility, coverage expectations, contract surface |
| `divisor-sre` | The Operator | File permissions/config, efficiency/performance, release pipeline [PACK], dependency health [PACK], runtime observability, upgrade paths, operational docs, backup/recovery | Deployment feasibility, operational requirements, config management, dependency risk, maintenance burden |
| `divisor-curator` | The Curator | Documentation gaps, blog/tutorial opportunities, website issue filing | Documentation completeness in specs, content coverage |

For any discovered agent not in this table, delegate with a generic review prompt appropriate to the current review mode.

---

## Code Review Mode

Review the current codebase for compliance with the Behavioral Constraints in `AGENTS.md`.

### Instructions

1. **Run local quality gates before delegating to
   council agents.** This step has three phases that
   MUST execute in order. All three phases apply only
   to Code Review Mode -- Spec Review Mode skips them.

   #### Phase 1a -- Pre-flight Checks (mandatory, hard gate)

   Load the `pre-flight` skill and run in `hard-gate`
   mode:

   a. Invoke the `skill` tool with name `pre-flight` to
      load the shared pre-flight check instructions.

   b. Execute the pre-flight skill's phases in order:
      1. CI Workflow Parsing — discover commands from
         `.github/workflows/`
      2. Local Tool Detection — check for config files
         and verify binary availability
      3. CI Coverage Matrix — display the matrix (in
         hard-gate mode, all tools are marked "Run
         locally = Yes")
      4. Execution — run all detected and available
         tools in hard-gate mode

   c. **If the pre-flight verdict is FAIL**: **STOP
      immediately.** Report each failure as a CRITICAL
      finding with the full error output. Do NOT
      proceed to Phase 1b or to step 2 (Divisor agent
      delegation). The rationale: reviewing code that
      doesn't compile or pass tests is wasted work.

   d. **If the pre-flight verdict is PASS**: report
      success and proceed to Phase 1b.

   #### Phase 1b -- Gaze Quality Analysis (conditional)

   a. Check if `gaze` is available:
      ```bash
      which gaze
      ```

   b. **If `gaze` is available**: invoke the
      `gaze-reporter` agent via the Task tool
      (subagent_type: `gaze-reporter`) with prompt
      `"full"` to produce a comprehensive quality
      report (CRAP scores, quality metrics,
      classification, health assessment). Capture
      the agent's output as the **Gaze Report**.

   c. **If `gaze` is NOT available**: skip with an
      informational note:
       > "Gaze not installed -- skipping quality
       > analysis. Install with
       > `brew install unbound-force/tap/gaze`
       > (or on Fedora/RHEL:
       > `go install github.com/unbound-force/gaze/cmd/gaze@latest`)."

      Proceed to step 2 without Gaze data.

   #### Phase 1c -- Discover Review Context (mandatory)

   Load the `review-context` skill for spec artifact
   discovery, path classification, and walkthrough
   generation:

   a. Invoke the `skill` tool with name `review-context`
      to load the shared context discovery instructions.

   b. Execute the skill's protocols in order:
      1. Protocol 1 (Spec Artifact Discovery) — locate
         the specification matching the current branch
         using the branch name from auto-detection and
         the changed file list.
      2. Protocol 2 (Issue Linking) — **conditional**.
         - If an **explicit PR number** was provided
           via `$ARGUMENTS` (see PR Number Argument
           above): fetch the PR body via
           `gh pr view <N> --json body --jq '.body'`
           and run Protocol 2 to extract linked issues
           and acceptance criteria. Pass the results
           to the Guard persona in Step 2 for
           concrete drift detection.
         - If **no explicit PR number** was provided:
           **skip**. Auto-detected PRs (from Step 7)
           are not available at Phase 1c time.
      3. Protocol 3 (Path-Based Focus Heuristics) —
         classify each changed file from the
         auto-detection step for review emphasis.
      4. Protocol 4 (Walkthrough Generation) — generate
         per-file change summaries from the branch
         diff (`git diff main...HEAD`).

   c. **Record results**: Use the skill's Review Context
      output format (Specification, File Classification,
      Walkthrough). This context is used in step 2
      (Divisor agent delegation) and step 6 (final
      report).

   d. **If the skill fails to load**: **STOP
      immediately.** Report the error as a CRITICAL
      finding. Do NOT proceed to step 2. The
      `review-context` skill is a hard dependency,
      consistent with the `pre-flight` skill
      consumption pattern — no inline fallback.

2. Delegate the review to all **discovered** reviewer agents in parallel using the Task tool. For each discovered agent, use the focus area from the Known Reviewer Roles reference table to provide targeted context. For any discovered agent not in the table, use a generic prompt: "Review the current changes for quality, correctness, and compliance. Return your verdict (APPROVE or REQUEST CHANGES) along with all findings."

   **CRITICAL — Review Scope Rule**: The review scope is
   ALWAYS the **full branch diff** (`git diff main...HEAD`),
   meaning ALL files changed on the branch relative to
   `main`. Do NOT narrow the scope to only recent commits,
   only uncommitted changes, or only files touched in the
   current session. Every agent MUST be instructed to read
   and review ALL changed files from the branch diff. The
   list of changed files from auto-detection step 2 MUST
   be included in each agent's prompt. Violating this rule
   produces incomplete reviews that miss findings in
   earlier commits on the branch.

   **Review context enrichment**: Append the following
   context sections to each Divisor agent's review
   prompt:

   - **Review Context** (from Phase 1c): Include the
     spec artifact summary, file classifications, and
     walkthrough from the `review-context` skill output.
     This gives agents spec alignment context and
     per-file focus heuristics. Instruct agents to
     reference spec requirements and file
     classifications in their findings where relevant.

   - **Quality Context** (from Phase 1b, when Gaze data
     is available): Include the Gaze Report summary.
     This gives agents -- particularly
     `divisor-testing` -- access to concrete CRAP
     scores, coverage percentages, quadrant
     distributions, and prioritized recommendations.
     Instruct agents to reference this data in their
     findings where relevant.

   **When Gaze data is NOT available**: include only
   the Review Context section. Agents review based on
   file reading plus spec/classification context.

   For each agent, instruct it to review the full branch diff (all changed files vs `main`) and return its verdict (**APPROVE** or **REQUEST CHANGES**) along with all findings.

3. Collect all **REQUEST CHANGES** findings from the
   discovered reviewers. If all discovered reviewers
   return **APPROVE**, report the result and stop.

   **Cross-persona finding consolidation**: Before
   proceeding to the fix loop, group findings from
   different personas that (a) affect the same
   component, file, or pipeline stage, (b) share a
   common root cause, and (c) together produce a risk
   greater than any individual finding. Merge each
   group into a single consolidated finding:
   - Apply compound severity escalation from
     `severity.md` to determine the combined severity.
   - Preserve per-persona attribution (e.g.,
     "Adversary: missing checksum + SRE: privileged
     blast radius → consolidated MEDIUM").
   - Present the consolidated finding with one unified
     recommendation addressing the root cause.

   Findings with independent root causes MUST remain
   separate even if they affect the same file.

4. If there are **REQUEST CHANGES**, address the findings by making the necessary code fixes. Then re-run all discovered reviewers to verify the fixes. Repeat this loop until all discovered reviewers return **APPROVE** or the process has exceeded 3 iterations.

5. If 3 iterations are exceeded, ask the user whether to continue or stop.

6. Provide a final report to the user:
   - **Discovery summary**: how many reviewer agents were discovered, which were invoked, and which known reviewer roles were absent (informational, non-blocking)
   - **Review context summary**: specification found
     (type, path) or "no spec found", and the
     walkthrough table from Phase 1c (review-context
     skill, Protocol 4)
   - What was found in each iteration
   - What was fixed
   - If stopped early, the current set of outstanding **REQUEST CHANGES**
   - If there were persistent circular **REQUEST CHANGES** (fixes for one reviewer cause failures in another), report those with additional detail so the user can make an informed decision

7. **GitHub Review Posting (optional, Code Review Mode only)**

   After the final report, offer to post the council's
   consolidated findings as a GitHub PR review. This step
   applies only to **Code Review Mode** — Spec Review
   Mode is a local pre-commit activity and does not post
   to GitHub. It is **opt-in** — it runs only when a PR
   exists and the user confirms posting.

   #### Step 7a -- PR Detection

   Detect whether the current branch has an open PR:

   a. If an **explicit PR number** was provided via
      `$ARGUMENTS`, use it directly. Skip auto-detection.

   b. Otherwise, attempt auto-detection:
      ```bash
      gh pr view --json number,headRefName,baseRefName
      ```
      If this succeeds, extract the PR number, head ref,
      and base ref name.

   c. **If `gh` is not installed**: skip Step 7 entirely
      with an informational note:
      > "GitHub CLI not available — skipping review
      > posting. Install `gh` for PR integration."

   c2. **If `gh` is installed but not authenticated**:
       verify with `gh auth status`. If authentication
       fails, skip Step 7 with:
       > "gh is installed but not authenticated —
       > skipping review posting. Run `gh auth login`
       > to enable PR integration."

   d. **If no PR exists** (auto-detection returns no PR
      and no explicit number provided): skip Step 7 with
      an informational note:
      > "No open PR found for this branch — review
      > remains local only."

   #### Step 7b -- Review State Fetching

   Fetch existing review state to prevent duplicate
   findings and enable pre-posting checks. Each sub-step
   is independent — if any fails, skip it and continue.

   **7b-i. Fetch Reviews**:
   ```bash
   gh api repos/{owner}/{repo}/pulls/<PR_NUMBER>/reviews \
     --jq '[.[] | {id: .id, user: .user.login, state: .state, body: .body, submitted_at: .submitted_at, commit_id: .commit_id}]'
   ```

   **7b-ii. Fetch Inline Comments**:
   ```bash
   gh api repos/{owner}/{repo}/pulls/<PR_NUMBER>/comments \
     --jq '[.[] | {path: .path, line: .line, body: .body, user: .user.login, created_at: .created_at}]'
   ```

   **7b-iii. Identify Current User**:
   ```bash
   gh api user --jq '.login'
   ```

   **7b-iv. Token Budget**: Cap existing review comments
   at 3000 characters total. When exceeded: filter to
   files changed in the branch diff, sort by `created_at`
   descending, include until budget exhausted, truncate
   remainder with a note.

   **7b-v. Error Handling**: If any `gh api` call returns
   403, 404, 429 (rate limited), or times out: log the
   error, skip the sub-step, proceed. All review state
   data is additive context — its absence reduces only
   deduplication accuracy.

   #### Step 7c -- Pre-posting Checks

   **Duplicate review detection**: Check if a review from
   the current user (7b-iii) already exists in the review
   list (7b-i):

   - If a prior review with the **same verdict** exists:
     Inform the user that a prior review exists and the
     latest review takes precedence. Use the
     **AskUserQuestion tool** with options
     `["Yes -- post new review", "No -- skip posting"]`.

   - If a prior review with a **different verdict** exists:
     Inform the user of the prior verdict and that the new
     review will override it. Use the
     **AskUserQuestion tool** with options
     `["Yes -- override with <new_verdict>",
     "No -- keep existing <old_verdict>"]`.

   - If no prior review exists: proceed silently.

   **Stale review + CODEOWNER checks** (APPROVE verdicts
   only): Fetch branch protection settings:

   ```bash
   gh api repos/{owner}/{repo}/branches/<baseRefName>/protection \
     --jq '{dismiss_stale: .required_pull_request_reviews.dismiss_stale_reviews, require_codeowners: .required_pull_request_reviews.require_code_owner_reviews}'
   ```

   If 404 (no branch protection) or 403 (insufficient
   permissions): skip both checks silently.

   If `dismiss_stale` is true:
   > "Warning: This repo dismisses stale reviews. If the
   > author pushes any new commits after this APPROVE, it
   > will be automatically invalidated and the PR will
   > return to REVIEW_REQUIRED. You may need to re-run
   > `/review-council` after final commits."

   If `require_codeowners` is true, check for CODEOWNERS:
   ```bash
   gh api repos/{owner}/{repo}/contents/CODEOWNERS \
     --jq '.name' 2>/dev/null || \
   gh api repos/{owner}/{repo}/contents/.github/CODEOWNERS \
     --jq '.name' 2>/dev/null
   ```
   If CODEOWNERS exists:
   > "Warning: This repo requires code owner reviews.
   > This APPROVE may not satisfy branch protection if
   > this account is not listed in CODEOWNERS."

   #### Step 7d -- Multi-Persona Finding Aggregation

   Assemble a single review body from all Divisor persona
   findings:

   **Review body structure**:

   ```
   ## Council Verdict: <VERDICT>

   **Reviewers**: <comma-separated persona names>
   **Iterations**: <count>

   ### <Persona Name> (<APPROVE | REQUEST CHANGES>)
   - [<SEVERITY>] <Finding description>
   - [<SEVERITY>] <Finding description>

   ### <Persona Name> (<APPROVE>)
   No findings.

   ...

   ---
   _This review was generated by /review-council
   (AI-assisted)._
   ```

   **Aggregation rules**:
   - LOW-severity findings: summarize as count only
     (e.g., "3 LOW findings omitted"). Do NOT enumerate
     each LOW finding in the review body.
   - MEDIUM+ findings: include full text with severity
     tag.
   - Consolidated cross-persona findings (from Step 3):
     present under the primary persona with attribution
     to contributing personas.
   - If the council verdict is **APPROVE WITH
     ADVISORIES**: include a note at the top:
     > "Council approved with advisories — unresolved
     > HIGH/CRITICAL findings require human judgment
     > before merge."

   **Body size limit**: If the assembled body exceeds
   60,000 characters, truncate per-persona sections
   starting from the persona with the most findings,
   replacing detailed findings with a summary count.
   Include: "Full findings available in the terminal
   report."

   #### Step 7e -- Inline Comment Preparation

   For findings mapped to specific files and line ranges
   in the diff, prepare inline comments:

   1. Collect all file-specific findings from all personas
   2. Sort by severity (CRITICAL > HIGH > MEDIUM > LOW)
   3. Within the same severity tier, round-robin across
      personas in **alphabetical order** by persona name.
      When the slot count is odd, the first persona
      alphabetically receives the extra slot.
   4. Take the top 15
   5. Overflow goes to the review body summary

   **Suggestion block format**: When a finding has a
   concrete single-file code fix (literal replacement):

   ````
   **[HIGH] Description of the issue**

   ```suggestion
   corrected code here
   ```
   ````

   Use suggestion blocks ONLY for literal code
   replacements. MUST NOT use them for architectural
   recommendations, multi-file changes, or removal of
   security controls.

   **Show all comments for review**: Present each comment
   to the user before posting:
   ```
   File: <path>
   Line: <line_number>
   Type: suggestion / plain-text
   Body: <comment text>
   ```

   #### Step 7f -- Verdict Mapping and Human Confirmation

   Map the council verdict to the GitHub API event type:

   | Council Verdict | GitHub Event |
   |-----------------|-------------|
   | APPROVE | `APPROVE` |
   | REQUEST CHANGES | `REQUEST_CHANGES` |
   | APPROVE WITH ADVISORIES | `COMMENT` |

   Display the verdict context, then use the
   **AskUserQuestion tool** for confirmation:

   For APPROVE verdicts:
   > "This will post an APPROVE review, which may unblock
   > merge in repos with branch protection. The review
   > will be labeled as AI-generated."

   Use options: `["Approve -- post review",
   "No -- skip posting", "Edit comments first",
   "Change verdict"]`.

   For REQUEST CHANGES or COMMENT verdicts:
   > "This will post a <verdict> review, which will
   > block merge in repos with branch protection."

   Use options: `["Yes -- post review",
   "No -- skip posting", "Edit comments first",
   "Change verdict"]`.

   - **"No -- skip posting"**: Skip posting. The terminal
     report is sufficient.
   - **"Edit comments first"**: Let the user modify
     comments, then re-confirm.
   - **"Change verdict"**: Let the user override the
     verdict (e.g., downgrade REQUEST CHANGES to
     COMMENT).

   **CRITICAL RULE**: NEVER post reviews without explicit
   human confirmation via the **AskUserQuestion tool**.
   Always show the exact content (verdict type + all
   comments) that will be posted and wait for the user
   to select a confirming option.

   #### Step 7g -- Post Review

   Construct a JSON payload containing:
   - `event`: the mapped GitHub event type
   - `body`: the assembled review body
   - `comments`: array of inline comment objects, each
     with `path` (file path relative to repo root),
     `line` (line number in the diff, right side), and
     `body` (comment text including severity tag)

   Write the payload to a temporary file and post:

   ```bash
   gh api repos/{owner}/{repo}/pulls/<PR_NUMBER>/reviews \
     --method POST \
     --input <json-file>
   ```

   Always write the JSON payload to a temporary file
   rather than interpolating into shell arguments, to
   prevent shell injection. Create the temporary file
   with restrictive permissions from the start — use
   `mktemp` (which creates files with mode 0600 on
   Linux) or set `umask 077` before creation to avoid
   a race window between creation and `chmod`.

   **Cleanup**: Remove the temporary file after posting,
   on ALL exit paths — including success, user
   cancellation, and posting failure.

   **Graceful degradation**: If `gh api` returns HTTP
   403, 404, or 422 (insufficient permissions, PR no
   longer exists, non-collaborator, or self-review
   prohibition):

   a. Fall back to posting as `"event": "COMMENT"` with
      a note:
      > "Note: Could not post as <original verdict> due
      > to insufficient permissions. Posted as COMMENT
      > instead. Original verdict: <verdict>."

   b. If the fallback also fails, inform the user that
      their token lacks write permissions for PR reviews
      and suggest re-authenticating with `gh auth login`.

   c. If `gh api` returns HTTP 429 (rate limited), skip
      posting with:
      > "GitHub API rate limit reached — posting skipped.
      > Retry later or post manually."

   **Auto-detected PR linked issues**: If the PR was
   auto-detected (not from explicit argument), parse the
   PR body (fetched during Step 7a) for issue references
   matching `Fixes #N`, `Closes #N`, or `Resolves #N`.
   For each matched reference:
   1. Validate the number is digits-only (1–999999)
   2. Fetch the issue: `gh issue view <N> --json title,body`
   3. Extract the title and any acceptance criteria
      (checkboxes or `## Acceptance Criteria` heading)
   4. Cap at 5 linked issues; truncate issue bodies at
      2000 characters

   Include the results as a "Linked Issues" section in
   the posted review body. This provides Protocol 2
   value for auto-detected PRs without requiring a
   Phase 1c re-run. If any `gh issue view` call fails,
   skip that issue silently.

---

## Spec Review Mode

Review spec artifacts for quality, consistency, and
alignment with the project constitution. The review scope
depends on the detected workflow tier.

### Determine Review Scope

Based on the workflow tier detected in the auto-detection
step, determine which artifacts to review:

- **Speckit** (branch `NNN-*`): Review the active spec
  directory at `specs/NNN-<name>/` (spec.md, plan.md,
  tasks.md, contracts/, data-model.md, checklists/),
  plus `.specify/memory/constitution.md` and `AGENTS.md`.

- **OpenSpec** (branch `opsx/*`): Review the active
  change directory at `openspec/changes/<name>/`
  (proposal.md, design.md, specs/, tasks.md), plus any
  referenced main specs at `openspec/specs/`, plus
  `.specify/memory/constitution.md` and `AGENTS.md`.

- **No active workflow** (main or unknown branch): Review
  all spec artifacts across both `specs/` and
  `openspec/specs/`, plus the constitution.

### Instructions

1. Delegate the review to all **discovered** reviewer agents in parallel using the Task tool. For each discovered agent, use the focus area from the Known Reviewer Roles reference table (selecting the Spec Review Focus column) to provide targeted context. For any discovered agent not in the table, use a generic prompt: "Review the spec artifacts in scope for quality, consistency, and alignment. Return your verdict (APPROVE or REQUEST CHANGES) along with all findings."

   For each agent, instruct it to **operate in Spec Review Mode**: review the spec artifacts identified in the review scope above (not code), plus `.specify/memory/constitution.md` and `AGENTS.md`. Include the workflow tier (Speckit/OpenSpec) in the agent prompt so it can tailor its review accordingly. Instruct the agent to return its verdict (**APPROVE** or **REQUEST CHANGES**) along with all findings.

2. Collect all **REQUEST CHANGES** findings from the
   discovered reviewers. If all discovered reviewers
   return **APPROVE**, report the result and stop.

   **Cross-persona finding consolidation**: Apply the
   same consolidation rule as Code Review Mode Step 3
   — group findings from different personas that share
   a root cause, apply compound severity escalation
   from `severity.md`, and present as consolidated
   findings with per-persona attribution preserved.

3. If there are **REQUEST CHANGES**, apply the **hybrid fix policy**:

   Severity levels are defined in the shared severity convention pack at `.opencode/uf/packs/severity.md`. The auto-fix boundary (LOW/MEDIUM = auto-fix, HIGH/CRITICAL = report only) is grounded in these shared definitions to ensure consistent behavior across all 5 personas.

   **Auto-fix (LOW and MEDIUM findings)** — Apply these fixes directly to the spec files:
   - Formatting and template compliance issues
   - Status field updates (e.g., "Draft" on a completed feature)
   - Terminology inconsistencies (same concept named differently across specs)
   - Missing or stale cross-references between spec, plan, and tasks
   - Coverage gaps with obvious fixes (e.g., a requirement with zero tasks when the task is clearly implied by the plan)
   - Stale or incorrect metadata (dates, branch names, prerequisite lists)

   **Report only (HIGH and CRITICAL findings)** — Do NOT attempt to fix these. Report them with full context and recommendations so the user can make an informed decision:
   - Missing user stories or acceptance criteria
   - Scope creep or under-specification
   - Design-level security gaps or unaddressed failure modes
   - Inter-feature conflicts or architectural misalignment
   - Constitution violations
   - Ambiguous requirements that require human judgment to resolve

4. After applying LOW/MEDIUM fixes, re-run all discovered reviewers to verify. Repeat this loop until all discovered reviewers return **APPROVE** (considering only remaining HIGH/CRITICAL findings as blocking) or the process has exceeded 3 iterations.

5. If 3 iterations are exceeded, ask the user whether to continue or stop.

6. Provide a final report to the user:
   - **Discovery summary**: how many reviewer agents were discovered, which were invoked, and which known reviewer roles were absent (informational, non-blocking)
   - What was found in each iteration
   - What was auto-fixed (LOW/MEDIUM)
   - Outstanding HIGH/CRITICAL findings that require human decision, with full context and recommendations
   - The Architect's Alignment Score for spec quality (if provided)
   - If there were persistent circular findings, report those with additional detail
   - Suggested next steps (e.g., "Run `/speckit.clarify` on spec 007 to resolve the ambiguous credential migration behavior")

---

## Verdict

The council returns **APPROVE** only when all discovered reviewers return **APPROVE**. Any single **REQUEST CHANGES** from a discovered reviewer means the council verdict is **REQUEST CHANGES**. Absent reviewers (known roles whose agent files were not found during discovery) do not affect the verdict but are noted in the discovery summary.

In Spec Review Mode, the council may return **APPROVE WITH ADVISORIES** when all LOW/MEDIUM findings have been auto-fixed but HIGH/CRITICAL findings remain that require human judgment. The advisories are the outstanding HIGH/CRITICAL findings. The discovery summary is included regardless of the verdict.
