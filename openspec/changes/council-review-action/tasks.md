# Council Review Action — Tasks

## 1. Action Packaging

- [x] 1.1 Create `council-review-action/action.yml` composite action
  with inputs (model, max-diff-lines, max-turns, max-budget-usd,
  claude-version, diff-path, meta-path, github-token, agents-pattern)
  and outputs (review-json, review-mode)
- [x] 1.2 Create `council-review-action/scripts/` directory for
  supporting scripts

## 2. Pre-fetch PR Context

- [x] 2.1 Create `prefetch.sh` — fetch CI check results via
  `gh pr checks` → `pr-checks.json`
- [x] 2.2 Fetch existing reviews via `gh api` → `pr-reviews.json`
  and inline comments → `pr-review-comments.json`
- [x] 2.3 Resolve linked issues from PR body (Fixes/Closes/Resolves
  #N) → `pr-linked-issues.json`, limit 5 issues, truncate bodies

## 3. Agent Discovery and Prompt Construction

- [x] 3.1 Create `discover-agents.py` — glob `divisor-*.md`, build
  `--agents` JSON via `json.dumps()`, output to file
- [x] 3.2 Create `build-prompt.sh` — construct review prompt from
  methodology file references (review-council.md, review-pr.md,
  severity.md, convention packs), CI constraints, JSON output schema
- [x] 3.3 Implement fallback flag when zero agents discovered

## 4. Claude Invocation and Output Parsing

- [x] 4.1 Create `run-review.sh` — invoke `claude -p` with
  `--agents` (multi-agent) or without (single-agent fallback),
  `--allowedTools`, budget/turn caps
- [x] 4.2 Parse output: validate JSON schema (summary +
  inline_comments), set review-mode output
- [x] 4.3 Handle Claude errors: separate stderr, log warnings,
  produce empty review JSON on failure

## 5. Validation

- [ ] 5.1 Verify `action.yml` is valid composite action syntax
- [ ] 5.2 Test agent discovery with unbound-force's own
  `.opencode/agents/divisor-*.md` files (9 personas)
- [ ] 5.3 Test single-agent fallback with no `divisor-*.md` files
- [ ] 5.4 End-to-end test via org-infra consuming the action
