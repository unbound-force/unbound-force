# Testing Strategy

## Local tests

Run from the `council-review-action` directory:

```bash
bash test/test-pipeline.sh
```

### Coverage matrix

| Script | Testable locally | Tests | Assertions |
|---|---|---|---|
| `prepare-diff.sh` | Yes | 8 | 31 |
| `filter-diff-lines.py` | Yes | 4 | 10 |
| `extract-review-json.py` | Yes | 8 | 8 |
| `build-prompt.sh` | Yes | 5 | 15 |
| `parse-output.sh` | Yes | 6 | 9 |
| `run-review.sh` | Yes (config generation via stub) | 3 | 17 |
| `divisor-*.md` | Yes (frontmatter syntax) | 1 | 1 |
| `prefetch.sh` | No (requires `gh` CLI + repo access) | — | — |

**Total: 35 scenarios, 91 assertions** (includes empty diff,
size limit, multiple JSON objects, 6 parse-output scenarios
covering direct JSON, JSONL streaming, plain text, filter
failure, and all-comments-filtered paths, JSON schema field
checks, parse attempt limit, multi-category noise filtering,
empty inline_comments, rename diff headers, special character
PR title handling, sandbox config generation for Vertex and
non-Vertex providers, MODEL metacharacter rejection, and
agent frontmatter permission syntax)

### Test details

#### prepare-diff.sh (noise filter + line annotation)

| # | Scenario | What it validates |
|---|---|---|
| 1 | New file | Lines annotated `[L1]` through `[LN]`, no annotation on hunk headers |
| 2 | Modified file | Context lines get correct new-file line numbers |
| 3 | Multi-file diff | Line numbers reset to 1 at each file boundary |
| 4 | Noise filtering | Lock/vendor/generated files excluded, code files kept |
| 5 | Deleted lines | `-` lines have no `[L]` prefix, line counter doesn't advance |
| 28 | Multi-category noise | Lock file, vendored, generated, and test fixture patterns all excluded |
| 30 | Rename headers | `similarity index`, `rename from/to` pass through without annotation |

#### filter-diff-lines.py (line validation)

| # | Scenario | What it validates |
|---|---|---|
| 6 | Valid vs invalid | Only lines within diff hunks accepted; invalid rescued to summary |
| 7 | Partial hunks | Lines outside the `@@` range rejected even if within file length |
| 29 | Empty comments | Empty `inline_comments` array passes through without summary modification |

#### extract-review-json.py (JSON extraction)

| # | Scenario | What it validates |
|---|---|---|
| 8 | Raw JSON | Direct JSON input parsed correctly |
| 9 | Code fences | JSON inside `` ```json `` fences extracted |
| 10 | Surrounding text | JSON embedded in prose extracted |
| 11 | Invalid input | Exit code 1 when no valid JSON found |
| 12 | Missing keys | JSON without `summary` + `inline_comments` rejected |
| 27 | Parse attempt limit | Input with many JSON fragments exceeding `MAX_PARSE_ATTEMPTS` rejected |

#### build-prompt.sh (prompt generation)

| # | Scenario | What it validates |
|---|---|---|
| 13 | PR title injection | Title appears in prompt, output format section present |
| 14 | Title truncation | Titles >200 chars are truncated |
| 15 | Security instructions | Untrusted input warning, no-shell/no-fix-loop constraints |
| 31 | Special characters | Shell metacharacters and backticks in PR title preserved literally |

#### run-review.sh (sandbox config)

| # | Scenario | What it validates |
|---|---|---|
| 32 | Vertex provider config | Permission denials (7 keys), task NOT denied, provider key present, --pure flag, no skip flags |
| 33 | Non-Vertex provider config | Permission denials present, no provider key |
| 34 | MODEL metachar rejection | Exit 1 for MODEL containing shell metacharacters |

#### divisor-*.md (agent frontmatter)

| # | Scenario | What it validates |
|---|---|---|
| 35 | Permission syntax | All divisor agents use `permission:` not `tools:` |

## Integration tests (live CI)

Scripts that require live credentials (`run-review.sh`, `prefetch.sh`) are validated through the CI workflow on evidence PRs.

### How to create an evidence PR

1. Push a branch with the workflow files to `org-infra`
2. Open a PR against `main`
3. The collect workflow runs on `pull_request`
4. Manually trigger the consumer workflow with the collect run ID:

```bash
gh workflow run ci_council_review.yml \
  --repo complytime/org-infra \
  --ref <branch> \
  -f triggering_run_id=<collect-run-id>
```

5. Verify:
   - Review summary posted as an issue comment
   - Inline comments on valid diff lines
   - Previous bot comments cleaned up

### What to check on the evidence PR

- All inline comments have line numbers within the file's actual line count
- No comments in "Additional findings (not on diff lines)" section (indicates line number mismatch)
- `review-mode` output is `inline` (not `comment` fallback)
- Stale bot comments from previous runs are deleted/minimized
