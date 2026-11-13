package coveragegate

import (
	"path"
	"sort"
	"strings"
)

// PackagesFromFiles derives module-relative package paths from a list of
// changed file paths. It drops test files, vendored code, common generated
// suffixes, non-Go files, and paths the go tool's package wildcard matching
// ignores, then deduplicates and sorts the resulting package directory set.
// The ignore rule keeps the changed-package set aligned with what a
// `go test ./...` coverprofile of this module can measure: wildcard-ignored
// paths appear in a git diff but can never appear in the profile, so
// demanding manifest coverage for them could never pass.
func PackagesFromFiles(files []string) []string {
	seen := make(map[string]struct{})
	var packages []string
	for _, file := range files {
		if !isProductionGoFile(file) || ignoredByGoTool(file) {
			continue
		}
		pkg := path.Dir(file)
		if _, ok := seen[pkg]; ok {
			continue
		}
		seen[pkg] = struct{}{}
		packages = append(packages, pkg)
	}
	sort.Strings(packages)
	return packages
}

// isProductionGoFile reports whether a module-relative path is a production Go
// source file (excluding tests, vendored code, and common generated suffixes).
func isProductionGoFile(file string) bool {
	if !strings.HasSuffix(file, ".go") {
		return false
	}
	if strings.HasSuffix(file, "_test.go") {
		return false
	}
	if strings.Contains(file, "/vendor/") || strings.HasPrefix(file, "vendor/") {
		return false
	}
	for _, generated := range []string{".pb.go", ".gen.go", "_generated.go"} {
		if strings.HasSuffix(file, generated) {
			return false
		}
	}
	return true
}

// ignoredByGoTool reports whether a path is ignored by the go tool's package
// wildcard matching. Per `go help packages`, directory and file names that
// begin with "." or "_" are ignored, as are directories named "testdata".
// Such paths show up in a git diff -- for example the .eval/ fixture trees
// that carry their own Go sources -- but never in the module's own test
// run, so the gate must not demand manifest coverage for them.
func ignoredByGoTool(file string) bool {
	segments := strings.Split(file, "/")
	for index, segment := range segments {
		if strings.HasPrefix(segment, ".") || strings.HasPrefix(segment, "_") {
			return true
		}
		if index < len(segments)-1 && segment == "testdata" {
			return true
		}
	}
	return false
}
