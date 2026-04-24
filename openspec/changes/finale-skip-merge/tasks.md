## 1. Remove Merge Step

- [x] 1.1 In `.opencode/command/finale.md`, remove the entire "### 7. Merge PR" section (the `gh pr merge` step and its error handling)
- [x] 1.2 Renumber "### 8. Return to Main" to "### 7. Return to Main"
- [x] 1.3 Renumber "### 9. Summary" to "### 8. Summary"

## 2. Update Summary Template

- [x] 2.1 In the Summary section of `.opencode/command/finale.md`, change the output template from reporting "merged via rebase" to "CI passed, ready for review". Add a "Next" line: "Request reviewers on the PR, then merge after approval with `gh pr merge --rebase --delete-branch`."

## 3. Update Guardrails

- [x] 3.1 In the Guardrails section of `.opencode/command/finale.md`, remove "NEVER merge with failing checks" (no longer applicable)
- [x] 3.2 Remove "ALWAYS use rebase merge" (no longer applicable)
- [x] 3.3 Add "NEVER merge the PR — /finale creates PRs for review, not for immediate merge"

## 4. Sync Scaffold Copy

- [x] 4.1 Copy the updated `.opencode/command/finale.md` to `internal/scaffold/assets/opencode/command/finale.md` (scaffold asset must match live file)

## 5. Validation

- [x] 5.1 Verify both files are identical: `diff .opencode/command/finale.md internal/scaffold/assets/opencode/command/finale.md`
- [x] 5.2 Run `go test -race -count=1 ./internal/scaffold/...` to verify scaffold asset count is unchanged (no new/removed files)
- [x] 5.3 Run `go build ./...` to verify build succeeds

<!-- spec-review: passed -->
<!-- code-review: passed -->
