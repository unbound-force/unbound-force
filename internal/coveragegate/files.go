package coveragegate

import (
	"path"
	"sort"
	"strings"
)

// PackagesFromFiles derives module-relative package paths from a list of
// changed file paths. It drops test files, vendored code, common generated
// suffixes, and non-Go files, then deduplicates and sorts the resulting
// package directory set.
func PackagesFromFiles(files []string) []string {
	seen := make(map[string]struct{})
	var packages []string
	for _, file := range files {
		if !isProductionGoFile(file) {
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
