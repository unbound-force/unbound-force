## 1. Implement Source Generation

- [x] 1.1 Add `generateDeweySources(opts *Options)` function to `internal/scaffold/scaffold.go`: scan `filepath.Dir(opts.TargetDir)` for sibling directories with `.git/`, extract GitHub org from `opts.ExecCmd("git", "remote", "get-url", "origin")`, generate multi-repo `sources.yaml` with per-repo disk sources + disk-org + GitHub API source. Only runs if `.dewey/sources.yaml` exists and has the default single-source content (check by counting `- id:` occurrences -- if > 1, user customized, skip).
- [x] 1.2 Add `isDefaultSourcesConfig(data []byte) bool` helper to `internal/scaffold/scaffold.go`: returns true if the file has exactly 1 `- id:` entry (the default from `dewey init`).
- [x] 1.3 Add `extractGitHubOrg(opts *Options) string` helper to `internal/scaffold/scaffold.go`: parses `git remote get-url origin` output. Handles SSH format (`git@github.com:ORG/repo.git`) and HTTPS format (`https://github.com/ORG/repo.git`). Returns empty string on failure (non-GitHub remote, no remote, exec error).
- [x] 1.4 Add `writeSourcesConfig(path, currentName string, siblings []string, parentDir, org string) error` helper: generates the YAML content with per-repo disk sources, disk-org, and optionally GitHub API source. Writes to the file.

## 2. Integrate into initSubTools

- [x] 2.1 Call `generateDeweySources(opts)` in `initSubTools()` in `internal/scaffold/scaffold.go`: insert after `dewey init` succeeds and before `dewey index`. Add the result to sub-tool results (e.g., "✅ Dewey sources: 5 repos detected" or "⊘ Dewey sources: already customized").

## 3. Tests

- [x] 3.1 Write `TestGenerateDeweySources_SiblingsDetected` in `internal/scaffold/scaffold_test.go`: create a temp dir structure with 3 sibling dirs containing `.git/`, run `generateDeweySources`, verify the written `sources.yaml` has per-repo disk sources + disk-org.
- [x] 3.2 Write `TestGenerateDeweySources_NoSiblings` in `internal/scaffold/scaffold_test.go`: create a temp dir with no siblings, verify sources.yaml has only disk-local + disk-org.
- [x] 3.3 Write `TestGenerateDeweySources_AlreadyCustomized` in `internal/scaffold/scaffold_test.go`: create a sources.yaml with 3 source entries, verify it is NOT overwritten.
- [x] 3.4 Write `TestExtractGitHubOrg_SSH` in `internal/scaffold/scaffold_test.go`: stub ExecCmd to return `git@github.com:unbound-force/repo.git`, verify org is `unbound-force`.
- [x] 3.5 Write `TestExtractGitHubOrg_HTTPS` in `internal/scaffold/scaffold_test.go`: stub ExecCmd to return `https://github.com/unbound-force/repo.git`, verify org is `unbound-force`.
- [x] 3.6 Write `TestExtractGitHubOrg_NonGitHub` in `internal/scaffold/scaffold_test.go`: stub ExecCmd to return a non-GitHub URL, verify org is empty string.
- [x] 3.7 Write `TestExtractGitHubOrg_NoRemote` in `internal/scaffold/scaffold_test.go`: stub ExecCmd to return error, verify org is empty string.
- [x] 3.8 Write `TestIsDefaultSourcesConfig` in `internal/scaffold/scaffold_test.go`: test with default content (1 source → true) and customized content (3 sources → false).

## 4. Verify

- [x] 4.1 Run `go build ./...`
- [x] 4.2 Run `go test -race -count=1 ./internal/scaffold/...`
- [x] 4.3 Run `go test -race -count=1 ./...`
