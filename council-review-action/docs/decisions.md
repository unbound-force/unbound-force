# Key Technical Decisions

## Two diff files: filtered vs annotated

**Decision**: Produce both `pr-diff-filtered.patch` (standard unified diff) and `pr-diff-annotated.patch` (with `[L<N>]` line prefixes).

**Why**: LLMs consistently confuse patch-file positions (sequential line count across the whole multi-file patch) with source-file line numbers (from `@@` hunk headers). For example, a 208-line file was getting comments on "line 243" because the model counted from the top of the concatenated patch. The `[L<N>]` prefix gives the model the correct number to read directly.

**Why two files**: `filter-diff-lines.py` needs standard unified diff format to parse `@@` hunk headers for validation. Annotated prefixes would break its parser. So the un-annotated version is used for validation, and the annotated version goes to the LLM.

**Pipeline**: `raw → filtered (noise removed) → annotated (line-numbered) → LLM`

## Three-workflow chain for fork PR support

**Decision**: Use collect + consumer + reusable instead of a single workflow.

**Why**: Fork PRs cannot access the base repo's secrets. The `pull_request` event fires in the fork's context. A `workflow_run` or `workflow_dispatch` trigger in the base repo runs with the base repo's secrets. The collect workflow (fork-safe, no secrets) uploads artifacts; the consumer downloads and authenticates.

**Alternative considered**: Single workflow with `pull_request_target` — rejected because it runs the workflow definition from `main`, not the PR branch, making it impossible to test workflow changes.

## Comment cleanup: delete + minimize

**Decision**: Delete issue comments and PR review comments, minimize Reviews API objects.

**Why**: The Reviews API (`POST /pulls/{n}/reviews`) creates objects that cannot be deleted via REST. GraphQL `minimizeComment` with classifier `OUTDATED` collapses them. Individual PR comments and issue comments can be deleted via REST.

**Alternative considered**: Supersede reviews with "Superseded by updated council review" body — rejected because it cluttered the PR timeline.

## continue-on-error on the review step

**Decision**: `continue-on-error: true` on the "Run council review" step in the caller workflow (`reusable_council_review.yml` in org-infra), not in this composite action.

**Why**: The council review is supplemental — a failure should not block CI. The downstream posting steps check `steps.review.outcome == 'success'` and skip if the review failed. This is the caller's responsibility, not the action's.

**Known gap**: No failure notification step yet. When the review fails silently, maintainers get no signal. A future improvement should add a step that posts a notice annotation on `outcome == 'failure'`.

## SHA-pinned action references

**Decision**: Pin the `council-review-action` to a full commit SHA, not a branch or tag.

**Why**: SHA-pinning is a supply-chain security best practice (scored by OSSF Scorecard). The SHA is immutable — even if the branch is rebased or force-pushed, the pinned commit stays the same.

**Current state**: Pinned to the `feat/council-review-action` feature branch SHA. Will be updated to a `main` SHA after the PR merges.

## Noise file filtering

**Decision**: Exclude lock files, vendored deps, generated code, test fixtures from the diff before review.

**Why**: These files add noise without review value. They inflate token costs and dilute review quality.

**Explicitly NOT excluded**: Spec files (openspec/, .specify/, docs/) — Divisor personas specifically review specs for intent drift and completeness.

## Prompt injection defense

**Decision**: Treat all diff content, PR titles, and file content as untrusted input. The prompt explicitly instructs the model to ignore override attempts.

**Scope**: This is a defense-in-depth measure. The primary defense is that the model runs in a sandboxed CI environment with no shell access, no network access beyond Vertex AI, and no file modification permissions beyond the review JSON output (OpenCode's `edit` permission covers `edit`, `write`, and `patch` tools).

## Runtime sandbox (defense-in-depth)

**Decision**: Enforce a three-layer runtime sandbox for the OpenCode review invocation:

1. **`OPENCODE_CONFIG_CONTENT` permission config** (hard boundary) — `run-review.sh` injects a runtime config via the `OPENCODE_CONFIG_CONTENT` env var that denies `edit`, `bash`, `webfetch`, `websearch`, and `skill`. Read, glob, and grep remain allowed (OpenCode defaults). This is enforced by the OpenCode permission system at runtime, not by prompt instruction. The permission config merges with the Vertex AI provider config already injected the same way.
2. **`--pure`** (plugin isolation) — External MCP plugins are not loaded, closing the plugin bypass vector where an external plugin could grant capabilities beyond the permission config.
3. **Agent frontmatter** (`permission:` in `divisor-*.md`) — Review-focused Divisor agents declare `edit: deny`, `bash: deny`, `webfetch: deny` in their frontmatter. This is defense-in-depth; even if the runtime config were misconfigured, the agent-level restrictions apply.

**Why `OPENCODE_CONFIG_CONTENT`**: OpenCode has no `--permissions` CLI flag. Permissions are configured via `opencode.json` or env var. `OPENCODE_CONFIG_CONTENT` is documented as "runtime overrides" at precedence level 6 (above project config) and merges with (not replaces) the project config.

**Why no `--dangerously-skip-permissions` or `--auto`**: These flags blanket-approve everything not explicitly denied. Instead, the permission config explicitly denies every tool that defaults to `"ask"` (`external_directory`, `doom_loop`) so no TTY prompt is needed. This is more precise — only the tools we've evaluated are allowed to run.

**Why `task` is NOT denied**: The orchestrator invokes Divisor subagents via the Task tool. OpenCode docs: "When set to deny, the subagent is removed from the Task tool description entirely, so the model won't attempt to invoke it." Denying `task` silently degrades to single-agent mode.

**Why not `--agent plan`**: The built-in Plan agent would replace Divisor persona discovery. `--agent plan` selects OpenCode's Plan agent instead of the custom Divisor agents, breaking multi-persona review.

**Why not `OPENCODE_EXPERIMENTAL_PLAN_MODE`**: Plan mode is "an instruction to the model, not a hard sandbox" (OpenCode docs). The permission config is a hard enforcement boundary, making plan mode redundant.

**References**: nunya#406 (defense-in-depth proposal), unbound-force#337 (original sandbox issue, closed as NOT_PLANNED — work folded into #406).

## OpenCode version pinning

**Decision**: Pin `opencode-ai@1.15.13` rather than using latest.

**Why**: Version `1.16.0+` changed the `--format json` streaming event structure — the `part.text` field used by `parse-output.sh` was removed. The JSONL parser (`grep '"type".*"text"' | jq '.part.text'`) requires the pre-1.16 format.

The pin was bumped from `1.2.26` to `1.15.13` because the runtime sandbox requires `--pure` (introduced in `1.3.5`). Version `1.2.26` did not have this flag.

**Safe range**: `1.3.5` through `1.15.13` (has `--pure` + `part.text` JSONL format). Versions `1.16.0+` require updating the JSONL parser before adoption.

## Pre-fetched context (full bodies)

**Decision**: Keep full review bodies, inline comments, and issue bodies in the pre-fetched context — no truncation.

**Why**: Human reviewer feedback is high-signal context. Truncating it causes the AI to duplicate findings or contradict prior feedback. The total token cost (~25-30K tokens) is well within the 200K context window.
